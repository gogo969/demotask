package promo

import "github.com/shopspring/decimal"

var (
	zero = decimal.NewFromInt(0)
)

//分页读取天天签到会员数量定义
const (
	SignPageSize      = 500  //每次拉取代理数量
	SignAwardPageSize = 1000 //每次拉取1000个签到会员帐号
)

const (
	PromoFloatingOpen  = 1
	PromoFloatingClose = 0
)

const (
	SignDefaultNum   = 4 //签到最小天数领取奖金
	SignDefaultAward = 0 //默认奖金
	SignAwardWayAuto = 3 //自动派奖
)

const (
	SignAwardWayManual = "1" //前台手动领取
)

type promoData struct {
	Period    uint8 `db:"period" json:"period"`           // 活动周期类型 0 指定周期 1 永久,
	DisplayAt int64 `db:"display_at" json:"display_time"` // 展示开始时间,
	HideAt    int64 `db:"hide_at" json:"hide_time"`       // 展示结束时间,
	StartAt   int64 `db:"start_at" json:"start_time"`     // 活动开始时间,
	EndAt     int64 `db:"end_at" json:"end_time"`         // 活动结束时间,
}

type Promo struct {
	ID                 string `db:"id" json:"id"`                       //
	Title              string `db:"title" json:"title"`                 // 活动标题,
	Flags              uint8  `db:"flags" json:"flags"`                 // 活动分类 61 全部、62 体育优惠、63 真人优惠、64 电竞优惠、65 彩票优惠、66 棋牌优惠、67 电子优惠,
	Tag                uint8  `db:"tag" json:"tag"`                     // 活动标签 71 无标签、72 最新、73 日常、74 新人、75 vip、76 限时、77 豪礼,
	Ty                 uint8  `db:"ty" json:"ty"`                       // 活动类型 41 纯展示模板、42 首存活动、43 邀请好友活动、44 天天签到活动、45 首存豪礼活动,
	ContentTy          uint16 `db:"content_ty" json:"content_ty"`       // 内容形式 421常规 422弹窗
	Sort               uint16 `db:"sort" json:"sort"`                   // 活动排序,
	Devices            string `db:"devices" json:"devices"`             // 支持端 二进制和 默认支持六端 00000001 web、 00000010 h5、 00000100 android、 00001000 ios、00010000 android_sport、 00100000 ios_sport,
	State              uint8  `db:"state" json:"state"`                 // 活动状态 0 关闭 1 开启,
	ShowState          uint8  `db:"-" json:"show_state"`                //
	Period             uint8  `db:"period" json:"period"`               // 活动周期类型 0 指定周期 1 永久,
	DisplayAt          int64  `db:"display_at" json:"display_time"`     // 展示开始时间,
	HideAt             int64  `db:"hide_at" json:"hide_time"`           // 展示结束时间,
	StartAt            int64  `db:"start_at" json:"start_time"`         // 活动开始时间,
	EndAt              int64  `db:"end_at" json:"end_time"`             // 活动结束时间,
	Images             string `db:"images" json:"images"`               // 页面设置图片,
	H5Images           string `db:"h5_images" json:"h5_images"`         // h5图片,
	ListImages         string `db:"-" json:"list_images"`               // web列表图
	BackGroundImages   string `db:"-" json:"back_ground_images"`        // web背景图
	HorizontalShareImg string `db:"-" json:"h_share_img"`               // 横屏分享图
	VerticalShareImg   string `db:"-" json:"v_share_img"`               // 竖屏分享图
	H5ListImages       string `db:"-" json:"h5_list_images"`            // 活动列表图
	H5ShareImages      string `db:"-" json:"h5_share_images"`           // 分享图
	Content            string `db:"content" json:"content"`             // 活动规则,
	H5Content          string `db:"h5_content" json:"h5_content"`       // h5活动规则,
	FloatingIcon       uint8  `db:"floating_icon" json:"floating_icon"` // 悬浮图标 ： 0 关闭 1 开启
	Description        string `db:"description" json:"description"`     // 活动说明{"desc":"%s","plat_desc":"%s","obj_desc":"%s"}
	CreatedAt          int64  `db:"created_at" json:"created_at"`       // 创建时间,
	UpdatedAt          int64  `db:"updated_at" json:"updated_at"`       // 更新时间,
	UpdatedUid         string `db:"updated_uid" json:"updated_uid"`     // 更新人id,
	UpdatedName        string `db:"updated_name" json:"updated_name"`   // 更新人name,
}

//SignRewardRecord 天天签到奖励派发记录表
type SignRewardRecord struct {
	Id             string  `db:"id" json:"id"`                             //
	Uid            string  `db:"uid" json:"uid"`                           // 会员id,
	Username       string  `db:"username" json:"username"`                 // 会员帐号,
	WeekStartAt    int64   `db:"week_start_at" json:"week_start_at"`       // 签到周周一时间戳,
	WeekEndAt      int64   `db:"week_end_at" json:"week_end_at"`           // 签到周周日时间戳,
	SignDayNum     uint8   `db:"sign_day_num" json:"sign_day_num"`         // 一周累计签到天数,
	ResignAt       int64   `db:"resign_at" json:"resign_at"`               // 一周补签时间
	SignDetail     string  `db:"sign_detail" json:"sign_detail"`           // 签到详情 0,0,0,0,0,0,0~1,1,1,1,1,1,1从周一至周日,每一位对应一天0表示未签到,1表示已签到,
	DepositAmount  float64 `db:"deposit_amount" json:"deposit_amount"`     // 一周累计存款,
	ValidBetAmount float64 `db:"valid_bet_amount" json:"valid_bet_amount"` // 一周累计有效投注,
	NetWinAmount   float64 `db:"net_win_amount" json:"net_win_amount"`     // 一周累计输赢,
	Amount         float64 `db:"amount" json:"amount"`                     // 应派奖金额,
	AwardWay       uint8   `db:"award_way" json:"award_way"`               // 派奖方式： 1前台手动领取、2后台审核派奖
	Resettle       uint8   `db:"resettle" json:"resettle"`                 // 是否二次派奖 0否 1是
	DividendId     string  `db:"dividend_id" json:"dividend_id"`           // 红利id
	HandOutState   uint16  `db:"hand_out_state" json:"hand_out_state"`     // 派奖状态：701未达标、702派奖成功、703派奖失败、704已失效、705已领取,
	HandOutAt      int64   `db:"hand_out_at" json:"hand_out_at"`           // 派奖时间,
	ReviewState    uint16  `db:"review_state" json:"review_state"`         // 审核状态：401待审核、402通过、403拒绝,
	ReviewAt       int64   `db:"review_at" json:"review_at"`               // 审核时间,
	CreatedAt      int64   `db:"created_at" json:"created_at"`             // 创建时间,
	ReviewUid      string  `db:"review_uid" json:"review_uid"`             // 审核人uid,
	ReviewName     string  `db:"review_name" json:"review_name"`           // 审核人,
	ReviewRemark   string  `db:"review_remark" json:"review_remark"`       // 审核备注,
	UpdatedAt      uint    `db:"updated_at" json:"updated_at"`             // 更新时间,
	UpdatedUid     string  `db:"updated_uid" json:"updated_uid"`           // 更新人id,
	UpdatedName    string  `db:"updated_name" json:"updated_name"`         // 更新人name,
}

// 签到记录
type SignRecord struct {
	ID             string  `db:"id" json:"id" redis:"id"`                                           // id
	UID            string  `db:"uid" json:"uid" redis:"uid"`                                        // uid
	Username       string  `db:"username" json:"username" redis:"username"`                         // 账号
	IsResign       uint8   `db:"is_resign" json:"is_resign" redis:"is_resign"`                      // 是否补签： 0 否 1 是
	Device         string  `db:"device" json:"device" redis:"device"`                               // 设备号
	SignAt         int64   `db:"sign_at" json:"sign_at" redis:"sign_at"`                            // 签到时间
	Day            int64   `db:"day" json:"day" redis:"day"`                                        // 签到日期
	DepositAmount  float64 `db:"deposit_amount" json:"deposit_amount" redis:"deposit_amount"`       // 充值金额
	ValidBetAmount float64 `db:"valid_bet_amount" json:"valid_bet_amount" redis:"valid_bet_amount"` // 有效投注
}

type signConfAward struct {
	WaterFlow string   `json:"water_flow"`
	Rules     []string `json:"rules"`
}

type signConfig struct {
	AwardWay       uint8           `json:"award_way"`        //派奖方式： 1前台手动领取、2后台审核派奖
	ResignWay      uint8           `json:"resign_way"`       //补签方式：1自动补签、2手动补签
	DepositAmount  string          `json:"deposit_amount"`   // 每日签到充值最小金额
	ValidBetAmount string          `json:"valid_bet_amount"` // 每日签到有效投注最小金额
	Award          []signConfAward `json:"award"`            //
}

type promoInvite struct {
	ID             string  `json:"id" db:"id"`
	DepositAmounts float64 `json:"deposit_amounts" db:"deposit_amounts"`
	GiftAmounts    float64 `json:"gift_amounts" db:"gift_amounts"`
}
