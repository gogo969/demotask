package dividend

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/beanstalkd/go-beanstalk"
	g "github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	cpool "github.com/silenceper/pool"
	"github.com/valyala/fasthttp"
	"lukechampine.com/frand"
	"strings"
	"task/contrib/conn"
	"task/contrib/helper"
	"task/modules/common"
	"task/party"
	"time"
)

var (
	beanPool      cpool.Pool
	platCfg       map[string]map[string]interface{}
	ctx           = context.Background()
	dialect       = g.Dialect("mysql")
	prefix        string
	lang          string
	merchantRedis *redis.Client
	merchantDB    *sqlx.DB
)

type dividendInfo struct {
	UID          string `db:"uid"`
	Username     string `db:"username"`
	Ty           int    `db:"ty"`
	Automatic    uint8  `db:"automatic"`
	PlatformID   string `db:"platform_id"`
	Wallet       uint8  `db:"wallet"`
	Amount       string `db:"amount"`
	HandOutState int    `db:"hand_out_state"`
}

func Parse(endpoints []string, path, topic string) {

	conf, platConf, err := common.ConfPlatParse(endpoints, path)
	if err != nil {
		fmt.Println("plat config parse error: ", err)
		return
	}

	prefix = conf.Prefix
	lang = conf.Lang
	merchantDB = conn.InitDB(conf.Db.Master.Addr, conf.Db.Master.MaxIdleConn, conf.Db.Master.MaxOpenConn)
	beanPool = conn.InitBeanstalk(conf.Beanstalkd.Addr, 15, 50, 100)
	merchantRedis = conn.InitRedisSentinel(conf.Redis.Addr, conf.Redis.Password, conf.Redis.Sentinel, conf.Redis.Db)

	// 场馆配置
	platCfg = platConf

	batchDividendTask(topic)
}

// 批量红利派发
func batchDividendTask(topic string) {

	common.Log("dividend", "红利派发服务开始")

	//循环监听
	for {
		v, err := beanPool.Get()
		if err != nil {
			fmt.Println("Beanstalk", 0, err.Error())
			_ = beanPool.Put(v)
			continue
		}

		if c, ok := v.(*beanstalk.Conn); ok {

			ts := beanstalk.NewTubeSet(c, topic)
			id, msg, err := ts.Reserve(2 * time.Minute)
			//无job时会返回timeout ,不打印日志，不处理
			if err != nil {

				if strings.Contains(err.Error(), "deadline soon") {
					//超时,续期
					_ = c.Touch(id)
				} else if !strings.Contains(err.Error(), "timeout") {
					fmt.Printf("tube: %s reserve error: %s\n", topic, err.Error())
				}

				_ = beanPool.Put(v)
				continue
			}

			message := string(msg)
			//记录日志
			fmt.Printf("tube: %s msg: %s running\n", topic, message)
			//避免过期，增加续约机制
			_ = c.Touch(id)

			// 获取参数
			param := map[string]interface{}{}
			m := &fasthttp.Args{}
			m.Parse(message)
			if m.Len() == 0 {
				fmt.Printf("tube: %s msg: %s parse error, deleted!\n", topic, message)
				_ = c.Delete(id)
				_ = beanPool.Put(v)
				continue
			}

			m.VisitAll(func(key, value []byte) {
				param[string(key)] = string(value)
			})

			dividendHandle(param)
			// 删除job
			_ = c.Delete(id)
		}

		_ = beanPool.Put(v)
	}
}

func dividendHandle(m map[string]interface{}) {

	common.Log("dividend", "dividend data : %v", m)

	data := dividendInfo{}

	now := time.Now()
	id := m["id"].(string)
	// 防止重复派发
	err := common.LockTTL(merchantRedis, "dvd:"+id, 1*time.Minute)
	if err != nil {
		common.Log("dividend", "duplicate order")
		return
	}

	ex := g.Ex{
		"id": id,
	}
	t := dialect.From("tbl_member_dividend")
	query, _, _ := t.Select("uid", "platform_id", "wallet", "username", "amount", "ty", "automatic", "hand_out_state").Where(ex).ToSQL()
	err = merchantDB.Get(&data, query)
	if err != nil {
		common.Log("dividend", "[%s] error: %v", query, err)
		return
	}

	if data.Username == "" {
		return
	}

	if data.HandOutState == common.DividendSuccess {
		common.Log("dividend", "duplicate order")
		return
	}

	dividend := g.Record{
		"id":            id,
		"uid":           data.UID,
		"username":      data.Username,
		"amount":        data.Amount,
		"review_remark": m["review_remark"],
		"review_at":     m["review_at"],
		"review_uid":    m["review_uid"],
		"review_name":   m["review_name"],
	}
	// 中心钱包红利
	if data.Wallet == 1 {

		cashType := common.TransactionDividend
		err = DividendHandOut(dividend, cashType)
		if err != nil {
			common.Log("dividend", "dividend hand out error : %v", err)
			return
		}

		return
	}

	ms := now.UnixMilli()
	// 场馆红利
	param := map[string]interface{}{
		"prefix":        prefix,
		"lang":          lang,
		"pid":           data.PlatformID,         //三方场馆的ID
		"transfer_type": common.TransferDividend, //场馆红利
		"deviceType":    "1",
		"tester":        "0", // 0 正式用户  1 测试用户
		"username":      data.Username,
		"uid":           data.UID,
		"id":            helper.GenId(),
		"ip":            common.GetClientIp(),
		"ms":            fmt.Sprintf("%d", ms),
		"s":             fmt.Sprintf("%d", now.Unix()),
		"amount":        data.Amount,
		"type":          "in",
		"confirm_at":    m["review_at"],
		"confirm_uid":   m["review_uid"],
		"confirm_name":  m["review_name"],
		"extend":        uint64(ms),
	}

	// im场馆北京时间处理
	//if lang == "vn" && (data.PlatformID == "5864536520308659696" || data.PlatformID == "1846182857231915191") {
	//	param["s8"] = fmt.Sprintf("%d", now.Unix()+3600)
	//	param["ms8"] = fmt.Sprintf("%d", now.UnixNano()/1e6+3600000)
	//}

	param["spid"] = param["pid"]
	pid := data.PlatformID
	if val, ok := party.PlatMap[data.PlatformID]; ok {
		pid = val
		param["pid"] = val
	}

	// 将场馆参数塞进param中
	for k, v := range platCfg[pid] {
		param[k] = v
	}

	param["prefix"] = prefix
	platInfo, err := platformRedis(data.PlatformID)
	if err != nil {
		platUnregisterMaintain(dividend)
		common.Log("dividend", "plat info error : %v", err)
		return
	}

	if platInfo["wallet"].(float64) == 0 {
		platUnregisterMaintain(dividend)
		common.Log("dividend", "platform maintain")
		return
	}

	_, ok := party.PlatBY[data.PlatformID]
	if ok {
		param["gamecode"] = platInfo["game_code"]
	}

	plat, _, err := common.MemberPlatformCache(merchantRedis, data.Username, pid)
	if err != nil || plat.ID == "" {

		plat, err = platReg(param)
		if err != nil {
			common.Log("dividend", "plat reg error : %v", err)
			platUnregisterMaintain(dividend)
			return
		}
	}

	param["password"] = plat.Password
	err = party.DividendPlatHandOut(merchantDB, merchantRedis, beanPool, param, dividend)
	if err != nil {
		common.Log("dividend", "plat dividend hand out error : %v", err)
		return
	}
}

func platReg(param map[string]interface{}) (common.MemberPlatform, error) {

	param["password"] = randPwd()
	code, _ := party.Dispatch("reg", param)
	if code == 0 {
		return common.MemberPlatform{}, errors.New(helper.PlatformRegErr)
	}

	mp, _, err := common.MemberPlatformInsert(merchantDB, merchantRedis, param)

	return mp, err
}

func platformRedis(pid string) (map[string]interface{}, error) {

	plat := map[string]interface{}{}
	k := fmt.Sprintf("plat:%s", pid)
	res, err := merchantRedis.Get(ctx, k).Result()
	if err == redis.Nil || err != nil {
		return plat, err
	}

	err = helper.JsonUnmarshal([]byte(res), &plat)
	if err != nil {
		return plat, err
	}

	return plat, nil
}

func platUnregisterMaintain(dividend g.Record) {

	record := g.Record{
		"state":          common.DividendReviewPass,
		"hand_out_state": common.DividendFailed,
		"review_remark":  dividend["review_remark"],
		"review_at":      dividend["review_at"],
		"review_uid":     dividend["review_uid"],
		"review_name":    dividend["review_name"],
	}
	ex := g.Ex{
		"id": dividend["id"],
	}
	// 更新调整记录状态
	query, _, _ := dialect.Update("tbl_member_dividend").Set(record).Where(ex).ToSQL()
	_, err := merchantDB.Exec(query)
	if err != nil {
		common.Log("dividend", "plat dividend hand out failed error : %v", err)
		return
	}
}

func randPwd() string {

	b := frand.Bytes(6)
	rp := hex.EncodeToString(b)
	return rp
}

// 中心钱包红利发放
func DividendHandOut(dividend g.Record, cashType int) error {

	var (
		balanceAfter decimal.Decimal
		balance      decimal.Decimal
	)

	id := dividend["id"].(string)
	uid := dividend["uid"].(string)
	name := dividend["username"].(string)
	amount := dividend["amount"].(string)

	money, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}

	// 获取中心钱包余额
	balance, err = common.MemberBalance(merchantDB, uid)
	if err != nil {
		return err
	}

	// 中心钱包转出
	balanceAfter = balance.Add(money)

	//1、判断金额是否合法
	if balanceAfter.IsNegative() {
		return errors.New(fmt.Sprintf("after amount : %s less than 0", balanceAfter.String()))
	}

	tx, err := merchantDB.Begin()
	if err != nil {
		return err
	}

	ex := g.Ex{
		"id": id,
	}
	record := g.Record{
		"state":           common.DividendReviewPass,
		"hand_out_state":  common.DividendSuccess,
		"hand_out_amount": amount,
		"review_remark":   dividend["review_remark"],
		"review_at":       dividend["review_at"],
		"review_uid":      dividend["review_uid"],
		"review_name":     dividend["review_name"],
	}
	// 更新调整记录状态
	query, _, _ := dialect.Update("tbl_member_dividend").Set(record).Where(ex).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	//4、新增账变记录
	trans := common.MemberTransaction{
		AfterAmount:  balanceAfter.String(),
		Amount:       amount,
		BeforeAmount: balance.String(),
		BillNo:       id,
		CreatedAt:    time.Now().UnixNano() / 1e6,
		ID:           helper.GenId(),
		CashType:     cashType,
		UID:          uid,
		Username:     name,
		Prefix:       prefix,
	}

	query, _, _ = dialect.Insert("tbl_balance_transaction").Rows(trans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	op := "+"
	// 红利金额为负数
	if money.IsNegative() {
		op = "-"
	}
	// 中心钱包上下分
	record = g.Record{
		"balance": g.L(fmt.Sprintf("balance%s%s", op, money.Abs().String())),
	}

	ex = g.Ex{
		"uid": uid,
	}
	query, _, _ = dialect.Update("tbl_members").Set(record).Where(ex).ToSQL()
	//fmt.Println(query)
	res, err := tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if r, _ := res.RowsAffected(); r == 0 {
		_ = tx.Rollback()
		return fmt.Errorf("affected 0 row: %s", query)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
