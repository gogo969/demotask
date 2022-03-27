package evo

import "github.com/shopspring/decimal"

const (
	PlatWalletLock   = "1" //场馆钱包锁定状态
	PlatWalletUnLock = "0" //场馆钱包解锁状态
)

//钱包类型
const (
	WalletTyDefault  = 0 //默认无钱包
	WalletTyBalance  = 1 //余额钱包
	WalletTyPlatform = 2 //场馆钱包
)

const (
	EVOID = "2306856765348772637"
)

const (
	DateF     = "2006-01-02 15:04:05"
	TDateF    = "2006-01-02T15:04:05"
	IBCFormat = "2006-01-02T15:04:05.999-04:00"
)

const (
	OperateAdd = "+"
	OperateSub = "-"
)

const EVO string = "EVO"

type EvoResult struct {
	Status   int8    `json:"status"`
	Type     int8    `json:"type"`
	GameId   int16   `json:"game_id"`
	Stake    float64 `json:"stake"`
	Payout   float64 `json:"payout"`
	IsCancel int8    `json:"is_cancel"`
	Error    string  `json:"error"`
}

//账变表
type MemberTransaction struct {
	AfterAmount  string `db:"after_amount"`  //账变后的金额
	Amount       string `db:"amount"`        //用户填写的转换金额
	BeforeAmount string `db:"before_amount"` //账变前的金额
	BillNo       string `db:"bill_no"`       //转账|充值|提现ID
	CashType     int    `db:"cash_type"`     //0:转入1:转出2:转入失败补回3:转出失败扣除4:存款5:提现
	CreatedAt    int64  `db:"created_at"`    //
	ID           string `db:"id"`            //
	UID          string `db:"uid"`           //用户ID
	Username     string `db:"username"`      //用户名
}

type BetTransaction struct {
	UID        string          `json:"uid"`         //会员id
	Username   string          `json:"username"`    //会员帐号
	BillNo     string          `json:"bill_no"`     //流水号
	Amount     decimal.Decimal `json:"amount"`      //帐变金额
	CashType   int             `json:"cash_type"`   //帐变类型
	CreateAt   int64           `json:"create_at"`   //操作时间
	PlatformID string          `json:"platform_id"` //场馆id
	WalletTy   int             `json:"wallet_ty"`   //钱包类型
	LockKey    string          `json:"lock_key"`    //锁定key
	Remark     string          `json:"remark"`      //三方流水号
}

type MemberBalanceData struct {
	Balance     decimal.Decimal //余额钱包金额
	PlatBalance decimal.Decimal //场馆钱包金额
	PlatState   string          //场馆钱包状态
	Total       decimal.Decimal //余额总计
}

type MemberPromoBalance struct {
	UID        string `db:"uid" json:"uid" redis:"uid"`                         //主键ID
	PlatformId string `db:"platform_id" json:"platform_id" redis:"platform_id"` //场馆id
	State      string `db:"state" json:"state" redis:"state"`                   //锁定状态
	LockAmount string `db:"lock_amount" json:"lock_amount" redis:"lock_amount"` //锁定余额
	WaterFlow  string `db:"water_flow" json:"water_flow" redis:"water_flow"`    //解锁总流水
	CreatedAt  string `db:"created_at" json:"created_at" redis:"created_at"`    //创建时间
}