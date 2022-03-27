package promo

import (
	"strconv"
	"task/modules/common"
)

// promoDividend 活动红利
func promoDividend(m map[string]interface{}) {

	// 获取锁定类型
	cashType, ok := m["cash_type"].(string)
	if !ok {
		common.Log("promo", "ty not found : %v", m)
		return
	}

	iTy, err := strconv.Atoi(cashType)
	if err != nil {
		common.Log("promo", "cash_type convert error: %s", err.Error())
		return
	}

	switch iTy {
	// 首存活动
	case common.TransactionFirstDepositDividend:
		firstDepositDividend(m)
	case common.TransactionSignDividend:
		flg, err := signDividendHandOut(iTy, m)
		if err != nil {
			common.Log("dividend-hand-out", "dividend-hand-out %s err:%s : %v", flg, err.Error(), m)
		}
	}
}
