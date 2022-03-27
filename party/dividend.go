package party

import (
	"fmt"
	"github.com/beanstalkd/go-beanstalk"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
	cpool "github.com/silenceper/pool"
	"github.com/valyala/fasthttp"
	"strconv"
	"task/contrib/helper"
	"task/modules/common"
	"time"
)

// 场馆红利发放
func DividendPlatHandOut(db *sqlx.DB, cli *redis.Client, mq cpool.Pool, param map[string]interface{}, dividend g.Record) error {

	id := param["id"].(string)
	amount := param["amount"].(string)
	dividendID := dividend["id"].(string)
	money, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}

	//1、判断金额是否合法
	if money.Cmp(minAmount) == -1 {
		return err
	}

	// 保留一位小数
	money = money.Truncate(1)
	if val, ok := transferScale[param["lang"].(string)][param["pid"].(string)]; ok {
		param["amount"] = money.Mul(decimal.NewFromInt(int64(val))).String()
	} else {
		param["amount"] = money.String()
	}

	// 场馆转账记录
	trans := g.Record{
		"id":            id,
		"uid":           param["uid"],
		"bill_no":       dividendID,
		"platform_id":   param["pid"],                // 三方场馆ID
		"username":      param["username"],           // 用户名
		"transfer_type": param["transfer_type"],      // 转账类型(0:转入1:转出2:后台上分3:场馆钱包清零)
		"amount":        amount,                      // 金额
		"before_amount": 0.0000,                      // 转账前的金额
		"after_amount":  0.0000,                      // 转账后的金额
		"created_at":    param["ms"],                 //
		"state":         common.TransferStateDealing, // 0:失败1:成功2:处理中3:脚本确认中4:人工确认中
		"automatic":     common.TransferConfirmAuto,  // 1:自动转账2:脚本确认3:人工确认
		"confirm_at":    param["confirm_at"],         // 更新时间
		"confirm_uid":   param["confirm_uid"],        // 操作人uid
		"confirm_name":  param["confirm_name"],       // 操作人名
	}
	mbTrans := g.Record{
		"id":            helper.GenLongId(),
		"after_amount":  0.0000,
		"amount":        amount,
		"before_amount": 0.0000,
		"bill_no":       dividendID,
		"created_at":    param["ms"],
		"cash_type":     common.TransactionPlatDividend,
		"uid":           param["uid"],
		"username":      param["username"],
	}

	// 场馆转账加锁
	err = common.LockTTL(cli, dividendID, 60*time.Minute)
	if err != nil {
		return err
	}

	ex := g.Ex{
		"id": dividendID,
	}
	record := g.Record{}
	// 向三方场馆转账
	r, c := Dispatch("transfer", param)
	fmt.Println("r: ", r, "dividend id : ", dividendID, "transfer id : ", param["id"], "pid : ", param["pid"], "amount: ", param["amount"], "c: ", c)
	switch r {
	case Success:
		// 转入成功，解锁
		common.Unlock(cli, dividendID)
		record = g.Record{
			"hand_out_state":  common.DividendSuccess, //发放成功
			"hand_out_amount": amount,
		}
		trans["bill_no"] = c
		trans["state"] = common.TransferStateSuccess //转账成功
	case Failure:
		// 转入成功，解锁
		common.Unlock(cli, dividendID)
		record = g.Record{
			"hand_out_state": common.DividendFailed, //发放失败
		}
		trans["state"] = common.TransferStateFailed //转账失败
	default:
		record = g.Record{
			"hand_out_state": common.DividendPlatDealing, //红利场馆上分处理中
		}
		param["transfer_type"] = fmt.Sprintf("%d", common.TransferDividend)
		param["dividend_id"] = dividendID
		param["hand_out_amount"] = amount
		fmt.Println(beanPut("platform", mq, param, 0))
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 更新调整记录状态
	query, _, _ := dialect.Update("tbl_member_dividend").Set(record).Where(ex).ToSQL()
	res, err := tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if r, _ := res.RowsAffected(); r == 0 {
		_ = tx.Rollback()
		return fmt.Errorf("affected 0 row : %s", query)
	}

	// 写入帐变表
	query, _, _ = dialect.Insert("tbl_balance_transaction").Rows(mbTrans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// 写入转账表
	query, _, _ = dialect.Insert("tbl_member_transfer").Rows(trans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// 上分成功，更新会员场馆状态
	if r == Success {

		record = g.Record{
			"transfer_in":            1,
			"transfer_in_processing": 0,
		}
		_, _ = common.MemberPlatformUpdate(db, cli, param["username"].(string), param["pid"].(string), record)
	}

	return nil
}

func Dispatch(name string, param map[string]interface{}) (int, string) {

	//EBET特殊处理一下
	if param["pid"].(string) == "2306868265751172637" && name == "confirm" {

		s, err := strconv.ParseInt(param["s"].(string), 10, 64)
		if err != nil {
			return Failure, helper.ParamErr
		}
		//如果超过2分钟则直接返回错误
		now := time.Now().Unix()
		if now-120 > s {
			return Failure, helper.ParamErr
		}
	}

	if rule, ok := rules[name]; ok {
		name1, ok1 := validator(rule, param)
		if !ok1 {
			fmt.Println("Dispatch Not Valid = ", param[name1])
			return Failure, name1 + " Not Valid"
		}
	} else {
		return Failure, "Not Found Method"
	}

	act := fmt.Sprintf("%s_%s", param["name"], name)
	if cb, ok := routes[act]; ok {
		return cb(param)
	}

	return Failure, "Not Found Method"
}

func beanPut(name string, mq cpool.Pool, param map[string]interface{}, delay int) (string, error) {

	m := &fasthttp.Args{}
	for k, v := range param {
		if _, ok := v.(string); ok {
			m.Set(k, v.(string))
		}
	}

	c, err := mq.Get()
	if err != nil {
		return "sys", err
	}

	if conn, ok := c.(*beanstalk.Conn); ok {

		tube := &beanstalk.Tube{Conn: conn, Name: name}
		_, err = tube.Put(m.QueryString(), 1, time.Duration(delay)*time.Second, 10*time.Minute)
		if err != nil {
			mq.Put(c)
			return "sys", err
		}
	}

	//将连接放回连接池中
	return "", mq.Put(c)
}
