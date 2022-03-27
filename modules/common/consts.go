package common

// 帐变类型
const (
	TransactionIn                    = 151 //场馆转入
	TransactionOut                   = 152 //场馆转出
	TransactionInFail                = 153 //场馆转入失败补回
	TransactionOutFail               = 154 //场馆转出失败扣除
	TransactionDeposit               = 155 //存款
	TransactionWithDraw              = 156 //提现
	TransactionUpPoint               = 157 //后台上分
	TransactionDownPoint             = 158 //后台下分
	TransactionDownPointBack         = 159 //后台下分回退
	TransactionDividend              = 160 //中心钱包红利派发
	TransactionRebate                = 161 //会员返水
	TransactionFinanceDownPoint      = 162 //财务下分
	TransactionWithDrawFail          = 163 //提现失败
	TransactionValetDeposit          = 164 //代客充值
	TransactionValetWithdraw         = 165 //代客提款
	TransactionAgencyDeposit         = 166 //代理充值
	TransactionAgencyWithdraw        = 167 //代理提款
	TransactionPlatUpPoint           = 168 //后台场馆上分
	TransactionPlatDividend          = 169 //场馆红利派发
	TransactionVIPUpgradeDividend    = 170 //vip升级红利
	TransactionFirstDepositDividend  = 171 //首存活动红利
	TransactionInviteDividend        = 172 //邀请好友红利
	TransactionBet                   = 173 //投注
	TransactionBetCancel             = 174 //投注取消
	TransactionPayout                = 175 //派彩
	TransactionResettlePlus          = 176 //重新结算加币
	TransactionResettleDeduction     = 177 //重新结算减币
	TransactionCancelPayout          = 178 //取消派彩
	TransactionPromoPayout           = 179 //场馆活动派彩
	TransactionEBetTCPrize           = 600 //EBet宝箱奖金
	TransactionEBetLimitRp           = 601 //EBet限量红包
	TransactionEBetLuckyRp           = 602 //EBet幸运红包
	TransactionEBetMasterPayout      = 603 //EBet大赛派彩
	TransactionEBetMasterRegFee      = 604 //EBet大赛报名费
	TransactionEBetBetPrize          = 605 //EBet投注奖励
	TransactionEBetReward            = 606 //EBet打赏
	TransactionEBetMasterPrizeDeduct = 607 //EBet大赛奖金取回
	TransactionWMReward              = 608 //WM打赏
	TransactionSBODividend           = 609 //SBO红利
	TransactionSBOReward             = 610 //SBO打赏
	TransactionSBOBuyLiveCoin        = 611 //SBO 购买LiveCoin
	TransactionSignDividend          = 612 //天天签到活动红利
	TransactionCQ9Dividend           = 613 //CQ9游戏红利
	TransactionCQ9PromoPayout        = 614 //CQ9活动派彩
	TransactionPlayStarPrize         = 615 //Playstar积宝奖金
	TransactionSpadeGamingRp         = 616 //SpadeGaming红包
	TransactionAEReward              = 617 //AE打赏
	TransactionAECancelReward        = 618 //AE取消打赏
	TransactionOfflineDeposit        = 619 //线下转卡存款
	TransactionUSDTOfflineDeposit    = 620 //USDT线下存款
	TransactionEVOPrize              = 621 //游戏奖金(EVO)
	TransactionEVOPromote            = 622 //推广(EVO)
	TransactionEVOJackpot            = 623 //头奖(EVO)
)

// 场馆转账类型
const (
	TransferIn           = 181 //场馆转入
	TransferOut          = 182 //场馆转出
	TransferUpPoint      = 183 //后台场馆上分
	TransferResetBalance = 184 //场馆钱包清零
	TransferDividend     = 185 //场馆红利
)

// 场馆转账状态
const (
	TransferStateFailed        = 191 //场馆转账失败
	TransferStateSuccess       = 192 //场馆转账成功
	TransferStateDealing       = 193 //场馆转账处理中
	TransferStateScriptConfirm = 194 //场馆转账脚本确认中
	TransferStateManualConfirm = 195 //场馆转账人工确认中
)

// 红利类型
const (
	DividendSite      = 211 //平台红利(站点)
	DividendUpgrade   = 212 //升级红利
	DividendBirthday  = 213 //生日红利
	DividendMonthly   = 214 //每月红利
	DividendRedPacket = 215 //红包红利
	DividendMaintain  = 216 //维护补偿
	DividendDeposit   = 217 //存款优惠
	DividendPromo     = 218 //活动红利
	DividendInvite    = 219 //推荐红利
	DividendAdjust    = 220 //红利调整
	DividendResetPlat = 221 //场馆余额负数清零
	DividendAgency    = 222 //代理红利
)

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

// 活动奖品审核状态
const (
	PromoGiftReviewing    = 401 //审核中
	PromoGiftReviewPass   = 402 //审核通过
	PromoGiftReviewReject = 403 //审核拒绝
)

// 首存活动记录状态
const (
	FDStateUnfinishedUncancelled = 451 //未完成未取消
	FDStateUnfinishedCancelled   = 452 //未完成已取消
	FDStateFinished              = 453 //已完成
)

// 活动状态变更类型
const (
	PromoAddWaterFlow = 461 //增加流水
	PromoUnlock       = 462 //解锁
)

// 天天签到派奖状态
const (
	SignPrizeUnReach        = 701 //未达标
	SignPrizeWaitHandOut    = 702 //待派奖
	SignPrizeHandOutSuccess = 703 //派奖成功
	SignPrizeHandOutFailed  = 704 //派奖失败
	SignPrizeInvalid        = 705 //已失效
	SignPrizeReceived       = 706 //已领取
)
