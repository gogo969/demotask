package common

// 红利审核状态
const (
	DividendReviewing    = 231 //红利审核中
	DividendReviewPass   = 232 //红利审核通过
	DividendReviewReject = 233 //红利审核不通过
)

// 红利发放状态
const (
	DividendFailed      = 236 //红利发放失败
	DividendSuccess     = 237 //红利发放成功
	DividendPlatDealing = 238 //红利发放场馆处理中
)

// 场馆转账确认类型
const (
	TransferConfirmAuto   = 241 //自动转账
	TransferConfirmScript = 242 //脚本确认
	TransferConfirmManual = 243 //人工确认
)

// 取款状态
const (
	WithdrawReviewing     = 371 //审核中
	WithdrawReviewReject  = 372 //审核拒绝
	WithdrawDealing       = 373 //出款中
	WithdrawSuccess       = 374 //提款成功
	WithdrawFailed        = 375 //出款失败
	WithdrawAbnormal      = 376 //异常订单
	WithdrawAutoPayFailed = 377 // 代付失败
	WithdrawHangup        = 378 // 挂起
	WithdrawDispatched    = 379 // 已派单
)
