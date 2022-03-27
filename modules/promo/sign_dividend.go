package promo

import (
	"errors"
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
	"task/contrib/helper"
	"task/modules/common"
	"time"
)

// signDividendHandOut 天天签到活动红利发放
func signDividendHandOut(cashType int, param map[string]interface{}) (string, error) {

	_, bonusAmount, id, mem, flg, err := signCheckParam(param)
	if err != nil {
		return flg, err
	}

	balance, balanceAfter, money, flg, err := signCheckBalance(bonusAmount, mem)
	if err != nil {
		return flg, err
	}

	divId := helper.GenId()
	record := signDividendRecord(divId, bonusAmount, mem)
	trans := signTransaction(divId, mem.UID, mem.Username, cashType, balanceAfter, money, balance)
	bRecord := g.Record{"balance": g.L(fmt.Sprintf("balance+%s", money.String()))}
	exMem := g.Ex{"uid": mem.UID}
	exRec := g.Ex{"id": id}

	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	// 更新调整记录状态
	query, _, _ := dialect.Insert("tbl_member_dividend").Rows(record).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	//4、新增账变记录
	query, _, _ = dialect.Insert("tbl_balance_transaction").Rows(trans).ToSQL()
	_, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	// 中心钱包上分
	query, _, _ = dialect.Update("tbl_members").Set(bRecord).Where(exMem).ToSQL()
	res, err := tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	if r, _ := res.RowsAffected(); r == 0 {
		_ = tx.Rollback()
		return "", fmt.Errorf("member:%s operateAmount:%s update balance failed", mem.UID, bonusAmount)
	}

	now := time.Now()
	//修改订单状态
	reward := g.Record{
		"hand_out_state": common.SignPrizeHandOutSuccess,
		"hand_out_at":    now.Unix(),
		"dividend_id":    divId,
	}
	query, _, _ = dialect.Update("tbl_promo_sign_reward_record").Set(reward).Where(exRec).ToSQL()
	res, err = tx.Exec(query)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}

	return "", nil
}

func signCheckParam(param map[string]interface{}) (string, string, string, common.Member, string, error) {

	username, ok := param["username"].(string)
	if !ok {
		return "", "", "", common.Member{}, "", errors.New(helper.ParamErr)
	}

	bonusAmount, ok := param["amount"].(string)
	if !ok {
		return "", "", "", common.Member{}, "", errors.New(helper.ParamErr)
	}

	id, ok := param["id"].(string)
	if !ok {
		return "", "", "", common.Member{}, "", errors.New(helper.ParamErr)
	}

	members, flg, err := common.MemberMCache(db, []string{username})
	if err != nil {
		return "", "", "", common.Member{}, flg, errors.New(helper.ParamErr)
	}

	if len(members) == 0 {
		return "", "", "", common.Member{}, "", errors.New(helper.ParamErr)
	}
	mem := members[username]

	var sid uint64
	exRec := g.Ex{
		"id":             id,
		"hand_out_state": []int{common.SignPrizeWaitHandOut, common.SignPrizeHandOutFailed},
	}
	query, _, _ := dialect.From("tbl_promo_sign_reward_record").Select("id").Where(exRec).ToSQL()
	err = db.Get(&sid, query)
	if err != nil {
		return username, bonusAmount, id, mem, "db", err
	}

	if sid == 0 {
		return username, bonusAmount, id, mem, "", errors.New(helper.GetDataFailed)
	}

	return username, bonusAmount, id, mem, "", nil
}

func signCheckBalance(bonusAmount string, mem common.Member) (decimal.Decimal, decimal.Decimal, decimal.Decimal, string, error) {

	var (
		balanceAfter decimal.Decimal
		balance      decimal.Decimal
	)
	money, err := decimal.NewFromString(bonusAmount)
	if err != nil {
		return zero, zero, zero, "", err
	}

	//1、判断金额是否合法
	if money.Cmp(zero) == -1 {
		return zero, zero, zero, "", errors.New(helper.AmountErr)
	}

	// 获取中心钱包余额
	balance, err = common.MemberBalance(db, mem.UID)
	if err != nil {
		return zero, zero, zero, "", err
	}

	// 中心钱包转出
	balanceAfter = balance.Add(money)

	return balance, balanceAfter, money, "", nil
}

func signTransaction(divId, uid, username string, cashType int, balanceAfter, money, balance decimal.Decimal) g.Record {

	return g.Record{
		"after_amount":  balanceAfter.String(),
		"amount":        money.String(),
		"before_amount": balance.String(),
		"bill_no":       divId,
		"created_at":    time.Now().UnixNano() / 1e6,
		"id":            helper.GenId(),
		"cash_type":     cashType,
		"uid":           uid,
		"username":      username,
		"prefix":        prefix,
	}
}

func signDividendRecord(id, bonusAmount string, mem common.Member) g.Record {

	now := time.Now()
	m := now.Unix()
	ms := now.UnixNano() / 1e6
	return g.Record{
		"id":              id,
		"uid":             mem.UID,
		"username":        mem.Username,
		"prefix":          prefix,
		"wallet":          1, // 1 中心钱包  2 场馆钱包
		"ty":              common.DividendPromo,
		"parent_uid":      mem.ParentUid,
		"parent_name":     mem.ParentName,
		"top_uid":         mem.TopUid,
		"top_name":        mem.TopName,
		"water_limit":     2, // 1 无需流水限制 2 需要流水限制'
		"water_flow":      bonusAmount,
		"amount":          bonusAmount,
		"remark":          "promo sign award",
		"apply_at":        ms,
		"apply_uid":       "0",
		"apply_name":      "promoSign",
		"batch":           1, //非批量发放
		"automatic":       0, //手动发放
		"batch_id":        "0",
		"platform_id":     "0",
		"hand_out_state":  common.DividendSuccess,
		"review_at":       m,
		"review_uid":      "0",
		"review_name":     "promoSign",
		"hand_out_amount": bonusAmount,
		"state":           common.DividendReviewPass,
	}
}
