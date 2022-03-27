package party

const tag string = "platform"

const (
	cn = "cn"
	vn = "vn"
)

const (
	Failure    int = 0 //失败
	Success    int = 1 //成功
	Pending    int = 3 //待处理
	Processing int = 2 //处理中
)

//账变表
type memberTransaction struct {
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

type PlatLog struct {
	Requesturl  string `json:"requestURL"`
	Requestbody string `json:"requestBody"`
	Statuscode  int    `json:"statusCode"`
	Body        string `json:"body"`
	Level       string `json:"level"`
	Err         string `json:"err"`
	Name        string `json:"name"`
	Username    string `json:"username"`
}

//场馆转账表
type memberTransfer struct {
	AfterAmount  string `db:"after_amount"`  //转账后的金额
	Amount       string `db:"amount"`        //金额
	Automatic    int    `db:"automatic"`     //1:自动转账2:脚本确认3:人工确认
	BeforeAmount string `db:"before_amount"` //转账前的金额
	BillNo       string `db:"bill_no"`       //
	CreatedAt    int64  `db:"created_at"`    //
	ID           string `db:"id"`            //
	PlatformID   string `db:"platform_id"`   //三方场馆ID
	State        int    `db:"state"`         //0:失败1:成功2:处理中3:脚本确认中4:人工确认中
	TransferType int    `db:"transfer_type"` //0:转入1:转出
	UID          string `db:"uid"`           //用户ID
	Username     string `db:"username"`      //用户名
	ConfirmAt    int64  `db:"confirm_at"`    //确认时间
	ConfirmUid   uint64 `db:"confirm_uid"`   //确认人uid
	ConfirmName  string `db:"confirm_name"`  //确认人名
}

var PlatMap = map[string]string{
	"7426646715018523638": "8318022162827355323", //CQ9捕鱼
	"6861705159854215439": "2658175169982643138", //AG捕鱼
	"934076801660754329":  "5864536520458745696", //DS棋牌
	"2854123669982643138": "5864536520458745696", //DS捕鱼
	"2854343669982643138": "5864536543458745696", //DS捕鱼(cn)
	"7426646735048523638": "8318022462867355323", //CQ9捕鱼(cn)
}

var PlatBY = map[string]bool{
	"8840968478572375866": true,
	"2854343669982643138": true,
	"6798530453614082003": true,
	"7426646735048523638": true,
}
