package promo

import (
	"fmt"
	"task/modules/common"

	g "github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
)

// 充值成功 发送红利
// param
// 		username 充值人的用户名
// 		amount  充值金额
//      deposit_created_at
// 		deposit_success_at
func depositSendDividendPromo(m map[string]interface{}) {

	username, ok := m["username"].(string)
	if !ok {
		common.Log("invite", "param username not found")
		return
	}

	successAt, ok := m["deposit_success_at"].(string)
	if !ok {
		common.Log("invite", "param deposit_success_at not found")
		return
	}

	// 被邀请人
	member, err := common.MemberFindOne(db, username)
	if err != nil {
		common.Log("invite", "child MemberCache err : %s", err.Error())
		return
	}

	amount, err := decimal.NewFromString(m["amount"].(string))
	if err != nil {
		common.Log("invite", "decimal.NewFromString err : %s, param: %#v", err.Error(), m)
		return
	}

	// 维护首存信息
	memberFirstDeposit(member, amount, successAt)
}

// 处理用户首存
func memberFirstDeposit(member common.Member, amount decimal.Decimal, successAt string) {

	if member.FirstDepositAt == 0 {
		// 更新用户首存金额 首存时间
		rec := g.Record{
			"first_deposit_at":     successAt,
			"first_deposit_amount": amount.Truncate(4).String(),
		}
		ex := g.Ex{
			"uid": member.UID,
		}
		query, _, _ := dialect.Update("tbl_members").Set(rec).Where(ex).ToSQL()
		fmt.Printf("memberFirstDeposit Update: %v\n", query)

		_, err := db.Exec(query)
		if err != nil {
			common.Log("invite", "update member error : %s, sql: %s", err.Error(), query)
			return
		}

		common.Log("invite", "member first deposit, username: %s, amount: %s", member.Username, amount.Truncate(4).String())
		return
	}
	fmt.Println(member.SecondDepositAt)
	if member.SecondDepositAt == 0 {
		// 更新用户首存金额 首存时间
		rec := g.Record{
			"second_deposit_at":     successAt,
			"second_deposit_amount": amount.Truncate(4).String(),
		}
		ex := g.Ex{
			"uid": member.UID,
		}
		query, _, _ := dialect.Update("tbl_members").Set(rec).Where(ex).ToSQL()
		fmt.Printf("memberSecondDeposit Update: %v\n", query)

		_, err := db.Exec(query)
		if err != nil {
			common.Log("invite", "update member error : %s, sql: %s", err.Error(), query)
			return
		}

		common.Log("invite", "member second deposit, username: %s, amount: %s", member.Username, amount.Truncate(4).String())
		return
	}
}
