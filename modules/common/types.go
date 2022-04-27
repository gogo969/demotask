package common

type Member struct {
	UID                string `db:"uid" json:"uid"`
	Username           string `db:"username" json:"username"`                         //会员名
	Password           string `db:"password" json:"password"`                         //密码
	RealnameHash       uint64 `db:"realname_hash" json:"realname_hash"`               //真实姓名哈希
	EmailHash          uint64 `db:"email_hash" json:"email_hash"`                     //邮件地址哈希
	PhoneHash          uint64 `db:"phone_hash" json:"phone_hash"`                     //电话号码哈希
	Prefix             string `db:"prefix" json:"prefix"`                             //站点前缀
	WithdrawPwd        uint64 `db:"withdraw_pwd" json:"withdraw_pwd"`                 //取款密码哈希
	Regip              string `db:"regip" json:"regip"`                               //注册IP
	RegDevice          string `db:"reg_device" json:"reg_device"`                     //注册设备号
	RegUrl             string `db:"reg_url" json:"reg_url"`                           //注册链接
	CreatedAt          uint32 `db:"created_at" json:"created_at"`                     //注册时间
	LastLoginIp        string `db:"last_login_ip" json:"last_login_ip"`               //最后登陆ip
	LastLoginAt        uint32 `db:"last_login_at" json:"last_login_at"`               //最后登陆时间
	SourceId           uint8  `db:"source_id" json:"source_id"`                       //注册来源 1 pc 2h5 3 app
	FirstDepositAt     uint32 `db:"first_deposit_at" json:"first_deposit_at"`         //首充时间
	FirstDepositAmount string `db:"first_deposit_amount" json:"first_deposit_amount"` //首充金额
	FirstBetAt         uint32 `db:"first_bet_at" json:"first_bet_at"`                 //首投时间
	FirstBetAmount     string `db:"first_bet_amount" json:"first_bet_amount"`         //首投金额
	TopUid             string `db:"top_uid" json:"top_uid"`                           //总代uid
	TopName            string `db:"top_name" json:"top_name"`                         //总代代理
	ParentUid          string `db:"parent_uid" json:"parent_uid"`                     //上级uid
	ParentName         string `db:"parent_name" json:"parent_name"`                   //上级代理
	BankcardTotal      uint8  `db:"bankcard_total" json:"bankcard_total"`             //用户绑定银行卡的数量
	LastLoginDevice    string `db:"last_login_device" json:"last_login_device"`       //最后登陆设备
	LastLoginSource    int    `db:"last_login_source" json:"last_login_source"`       //上次登录设备来源:1=pc,2=h5,3=ios,4=andriod
	Remarks            string `db:"remarks" json:"remarks"`                           //备注
	State              uint8  `db:"state" json:"state"`                               //状态 1正常 2禁用
	Balance            string `db:"balance" json:"balance"`                           //余额
	LockAmount         string `db:"lock_amount" json:"lock_amount"`                   //锁定金额
	Commission         string `db:"commission" json:"commission"`                     //佣金
}

// MemberPlatform 会员场馆表
type MemberPlatform struct {
	ID                    string `db:"id" json:"id" redis:"id"`                                                                //
	Username              string `db:"username" json:"username" redis:"username"`                                              //用户名
	Pid                   string `db:"pid" json:"pid" redis:"pid"`                                                             //场馆ID
	Password              string `db:"password" json:"password" redis:"password"`                                              //平台密码
	Balance               string `db:"balance" json:"balance" redis:"balance"`                                                 //平台余额
	State                 int    `db:"state" json:"state" redis:"state"`                                                       //状态:1=正常,2=锁定
	CreatedAt             uint32 `db:"created_at" json:"created_at" redis:"created_at"`                                        //
	TransferIn            int    `db:"transfer_in" json:"transfer_in" redis:"transfer_in"`                                     //0:没有转入记录1:有
	TransferInProcessing  int    `db:"transfer_in_processing" json:"transfer_in_processing" redis:"transfer_in_processing"`    //0:没有转入等待记录1:有
	TransferOut           int    `db:"transfer_out" json:"transfer_out" redis:"transfer_out"`                                  //0:没有转出记录1:有
	TransferOutProcessing int    `db:"transfer_out_processing" json:"transfer_out_processing" redis:"transfer_out_processing"` //0:没有转出等待记录1:有
	Extend                uint64 `db:"extend" json:"extend" redis:"extend"`                                                    //兼容evo
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
	Prefix       string `db:"prefix"`        //站点前缀
}

type MBBalance struct {
	UID        string `db:"uid" json:"uid"`
	Balance    string `db:"balance" json:"balance"`         //余额
	LockAmount string `db:"lock_amount" json:"lock_amount"` //锁定额度
	Commission string `db:"commission" json:"commission"`   //代理余额
}

// 站内信
type Message struct {
	MsgID    string `json:"msg_id"`    //站内信id
	Username string `json:"username"`  //会员名
	Title    string `json:"title"`     //标题
	SubTitle string `json:"sub_title"` //标题
	Content  string `json:"content"`   //内容
	IsTop    string `json:"is_top"`    //0不置顶 1置顶
	IsVip    string `json:"is_vip"`    //0非vip站内信 1vip站内信
	Ty       string `json:"ty"`        //1站内消息 2活动消息
	SendName string `json:"send_name"` //发送人名
	SendAt   int64  `json:"send_at"`   //发送时间
	Prefix   string `json:"prefix"`    //商户前缀
}
