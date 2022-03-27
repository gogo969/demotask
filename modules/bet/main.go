package bet

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/panjf2000/ants/v2"
	cpool "github.com/silenceper/pool"
	"task/contrib/conn"
	"task/contrib/helper"
	"task/modules/common"
	"time"
)

var (
	betDB    *sqlx.DB
	beanPool cpool.Pool
	dialect  = g.Dialect("mysql")
)

type gameRecord struct {
	RowID          string `db:"row_id" json:"row_id"`
	BillNO         string `db:"bill_no" json:"bill_no"`
	ApiType        string `db:"api_type" json:"api_type"`
	PlayerName     string `db:"player_name" json:"player_name"`
	Name           string `db:"name" json:"name"`
	UID            string `db:"uid" json:"uid"`
	NetAmount      string `db:"net_amount" json:"net_amount"`
	BetTime        string `db:"bet_time" json:"bet_time"`
	GameType       string `db:"game_type" json:"game_type"`
	BetAmount      string `db:"bet_amount" json:"bet_amount"`
	ValidBetAmount string `db:"valid_bet_amount" json:"valid_bet_amount"`
	Flag           string `db:"flag" json:"flag"`
	PlayType       string `db:"play_type" json:"play_type"`
	Prefix         string `db:"prefix" json:"prefix"`
	Result         string `db:"result" json:"result"`
	CreatedAt      string `db:"created_at" json:"created_at"`
	UpdatedAt      string `db:"updated_at" json:"updated_at"`
	ApiName        string `db:"api_name" json:"api_name"`
	ApiBillNO      string `db:"api_bill_no" json:"api_bill_no"`
	MainBillNO     string `db:"main_bill_no" json:"main_bill_no"`
	GameName       string `db:"game_name" json:"game_name"`
	GameCode       string `db:"game_code" json:"game_code"`
	HandicapType   string `db:"handicap_type" json:"handicap_type"`
	Handicap       string `db:"handicap" json:"handicap"`
	Odds           string `db:"odds" json:"odds"`
	SettleTime     string `db:"settle_time" json:"settle_time"`
	StartTime      string `db:"start_time" json:"start_time"`
	Resettle       string `db:"resettle" json:"resettle"`
	Presettle      string `db:"presettle" json:"presettle"`
	ApiBetTime     string `db:"api_bet_time" json:"api_bet_time"`
	ApiSettleTime  string `db:"api_settle_time" json:"api_settle_time"`
	ParentUID      string `db:"parent_uid" json:"parent_uid"`
	ParentName     string `db:"parent_name" json:"parent_name"`
	RebateAmount   string `db:"rebate_amount" json:"rebate_amount"`
	TopUID         string `db:"top_uid" json:"top_uid"`
	TopName        string `db:"top_name" json:"top_name"`
}

func Parse(endpoints []string, path string) {

	conf := common.ConfParse(endpoints, path)
	// 初始化beanstalk
	beanPool = conn.InitBeanstalk(conf.Beanstalkd.Addr, 50, 50, 100)
	// 初始化db
	betDB = conn.InitDB(conf.Db.Bet.Addr, conf.Db.Bet.MaxIdleConn, conf.Db.Bet.MaxIdleConn)

	batchBetTask()
}

// 投注
func batchBetTask() {

	// 初始化投注记录任务队列协程池
	confirmPool, _ := ants.NewPoolWithFunc(500, func(bet interface{}) {

		if fn, ok := bet.(common.BeansFnParam); ok {
			// 场馆转账订单确认
			err := betHandle(fn.M)
			if err == nil {
				// 删除job
				_ = fn.Conn.Delete(fn.ID)
			}
		}
	})

	attr := common.BeansWatcherAttr{
		TubeName:       "bet",
		ReserveTimeOut: 2 * time.Minute,
		Pool:           confirmPool,
	}

	// 投注记录确认队列
	common.BeanstalkWatcher(beanPool, attr)
}

// map 参数
func betHandle(param map[string]interface{}) error {

	if _, ok := param["row_id"].(string); !ok {
		fmt.Println("betHandle row_id null param = ", param)
		return errors.New("no row_id")
	}

	state, ok := param["flag"].(string)
	if !ok {
		fmt.Println("betHandle param = ", param)
		return errors.New("no flag")
	}

	//为了处理取消派彩等状态
	if state == "0" && len(param) > 30 {
		//新增
		return insertBet(param)
	}

	//更新
	return updateBet(param)
}

//新增注单
func insertBet(param map[string]interface{}) error {

	jsonB, err := helper.JsonMarshal(param)
	if err != nil {
		fmt.Println("insertBet jsonB err = ", err)
		return err
	}

	data := gameRecord{}
	err = helper.JsonUnmarshal(jsonB, &data)
	if err != nil {
		fmt.Println("insertBet data err = ", err)
		return err
	}

	if len(data.RebateAmount) == 0 {
		data.RebateAmount = "0"
	}

	query, _, _ := dialect.Insert("tbl_game_record").Rows(data).ToSQL()
	_, err = betDB.Exec(query)
	if err != nil {
		fmt.Println("insertBet query = ", query)
		return err
	}

	return nil
}

//更新注单
func updateBet(param map[string]interface{}) error {

	paramLen := len(param)
	if paramLen < 1 {
		return errors.New("param len")
	}

	rec := g.Record{}
	for k, v := range param {

		if k == "prefix" {
			continue
		}

		if k == "bet_amount" && paramLen == 3 {
			rec[k] = g.L(fmt.Sprintf("bet_amount+%s", v.(string)))
			continue
		}

		rec[k] = v
	}

	query, _, _ := dialect.Update("tbl_game_record").Where(g.Ex{"row_id": param["row_id"].(string)}).Set(rec).ToSQL()
	res, err := betDB.Exec(query)
	if err != nil {
		fmt.Println("updateBet err = ", err)
		fmt.Println("updateBet query = ", query)
		return err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return errors.New("no update")
	}

	return nil
}
