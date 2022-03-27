package promo

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/olivere/elastic/v7"
	"github.com/valyala/fastjson"
	"task/contrib/helper"
	"task/modules/common"
)

//signReport 签到有效投注更新
//注意:定时执行时，在每日00:00:00~00:30:00需要对前一天数据进行更新
func signReport(day int64) {

	startDay := helper.DayTST(day, loc).Unix()
	ex := g.Ex{
		"day":       startDay,
		"is_resign": 0,
		"prefix":    prefix,
	}

	total := signMemberTotal(ex)
	pages := total / SignPageSize
	if total%SignPageSize != 0 {
		pages = pages + 1
	}

	var page int64
	for ; page < pages; page++ {

		query, _, _ := dialect.From("tbl_promo_sign_record").
			Select("username").Where(ex).Offset(uint(page * SignPageSize)).Limit(SignPageSize).ToSQL()
		var users []string
		err := db.Select(&users, query)
		if err != nil {
			return
		}

		fields := []string{"username", "valid_bet_amount", "deposit_amount"}
		boolQuery := elastic.NewBoolQuery().Filter(
			elastic.NewTermsQueryFromStrings("username", users...),
			elastic.NewTermQuery("report_time", startDay),
		)
		t, data, _, flg, err := common.EsQuerySearch(reportEs, pullPrefix+"t_mem_settle_finance_report",
			"username", 1, SignPageSize, fields, boolQuery, nil)
		if err != nil {
			common.Log("sign", "sign_report %s err: %s", flg, err.Error())
			return
		}

		if t == 0 {
			continue
		}

		var p fastjson.Parser
		for _, v := range data {

			body, err := v.Source.MarshalJSON()
			if err != nil {
				common.Log("sign", "sign_report err:%s", err.Error())
				continue
			}

			bucket, err := p.ParseBytes(body)
			if err != nil {
				common.Log("sign", "sign_report err:%s", err.Error())
				continue
			}

			uex := g.Ex{
				"day":       startDay,
				"username":  string(bucket.GetStringBytes("username")),
				"is_resign": 0,
			}
			record := g.Record{
				"valid_bet_amount": fmt.Sprintf("%0.4f", bucket.GetFloat64("valid_bet_amount")),
				"deposit_amount":   fmt.Sprintf("%0.4f", bucket.GetFloat64("deposit_amount")),
			}
			uQuery, _, _ := dialect.Update("tbl_promo_sign_record").Where(uex).Set(record).ToSQL()
			_, err = db.Exec(uQuery)
			if err != nil {
				common.Log("update", "query:%s err:%s", uQuery, err.Error())
			}
		}
	}
}

//commissionAgencyCount 代理总条数
func signMemberTotal(ex g.Ex) int64 {

	var total int64
	countQuery, _, _ := dialect.From("tbl_promo_sign_record").Select(g.COUNT("id")).Where(ex).ToSQL()
	err := db.Get(&total, countQuery)
	fmt.Println(countQuery)
	if err != nil || total < 1 {
		return 0
	}

	return total
}
