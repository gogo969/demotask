package promo

import (
	g "github.com/doug-martin/goqu/v9"
	"strconv"
	"task/modules/common"
	"time"
)

func firstDepositDividend(m map[string]interface{}) {

	now := time.Now()
	// 增加流水 / 解锁
	alterTy, ok := m["alter_ty"].(string)
	if !ok {
		common.Log("promo", "alter_ty not found : %v", m)
		return
	}

	ty, err := strconv.Atoi(alterTy)
	if err != nil {
		common.Log("promo", "alter_ty convert: %v", m)
		return
	}

	// 解锁
	if ty == common.PromoUnlock {

		ex := g.Ex{
			"id":    m["pid"],
			"state": common.FDStateUnfinishedUncancelled,
		}
		rec := g.Record{
			"state":       common.FDStateFinished,
			"unlock_at":   now.Unix(),
			"unlock_ty":   m["unlock_ty"], // 流水达标解锁  余额解锁
			"finished_at": now.Unix(),
		}
		query, _, _ := dialect.Update("tbl_promo_first_deposit_record").Set(rec).Where(ex).ToSQL()
		_, err = db.Exec(query)
		if err != nil {
			common.Log("promo", "update error: %v, sql: %s", err.Error(), query)
		}
	}

	// 添加流水
	if ty == common.PromoAddWaterFlow {

		ex := g.Ex{
			"id":    m["pid"],
			"state": common.FDStateUnfinishedUncancelled,
		}

		rec := g.Record{
			"additional_water": g.L("additional_water + ?", m["water_flow"].(string)),
		}
		query, _, _ := dialect.Update("tbl_promo_first_deposit_record").Set(rec).Where(ex).ToSQL()
		_, err = db.Exec(query)
		if err != nil {
			common.Log("promo", "update error: %v, sql: %s", err.Error(), query)
		}
	}
}
