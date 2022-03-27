package promo

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/go-redis/redis/v8"
	"github.com/olivere/elastic/v7"
	"github.com/shopspring/decimal"
	"github.com/valyala/fastjson"
	"strconv"
	"strings"
	"task/contrib/helper"
	"task/modules/common"
	"time"
)

//signHandoutAward 天天签到入口
func signHandoutAward(day int64) {

	lastWeekStartAt := helper.WeekTST(day, loc).AddDate(0, 0, -7).Unix()
	lastWeekEndAt := helper.WeekTET(day, loc).AddDate(0, 0, -7).Unix()

	conf, flg, err := signLoadConf()
	if err != nil {
		common.Log("promo-sign-award-config", "promo_sign_award_config %s err: %s", flg, err.Error())
		return
	}

	total, flg, err := signEsMemberTotal(lastWeekStartAt, lastWeekEndAt)
	if err != nil {
		common.Log("promo-sign-total", "es members_total %s err: %s", flg, err.Error())
		return
	}

	pages := total / SignAwardPageSize
	if total%SignAwardPageSize != 0 {
		pages = pages + 1
	}
	var page int64
	for ; page < pages; page++ {

		users, userList, flg, err := signMembers(lastWeekStartAt, lastWeekEndAt, page)
		if err != nil {
			common.Log("promo-sign-reward", "es members %s err: %s", flg, err.Error())
			continue
		}

		fmt.Printf("users :%v\n", users)
		records, flg, err := signMemberReward(lastWeekStartAt, lastWeekEndAt, users, userList, conf)
		if err != nil {
			common.Log("promo-sign-reward", "generate records %s err: %s", flg, err.Error())
			continue
		}

		fmt.Printf("users :%v\n", records)
		if len(records) > 0 {
			query, _, _ := dialect.Insert("tbl_promo_sign_reward_record").Rows(records).ToSQL()
			fmt.Printf("insert reward :%v\n", query)
			_, err = db.Exec(query)
			if err != nil {
				common.Log("promo-sign-reward-insert", "add records to query: %s err: %s", query, err.Error())
				continue
			}
		}

		signClearCache(lastWeekStartAt, users)
	}

	//处理超过7天未领取奖金记录，将状态更新为失效
	invalidWeekAt := helper.WeekTST(day, loc).AddDate(0, 0, -14).Unix()
	ex := g.Ex{
		"week_start_at":  invalidWeekAt,
		"hand_out_state": []int{common.SignPrizeWaitHandOut, common.SignPrizeHandOutFailed},
	}
	query, _, _ := dialect.Update("tbl_promo_sign_reward_record").
		Set(g.Record{"hand_out_state": common.SignPrizeInvalid}).Where(ex).ToSQL()
	_, err = db.Exec(query)
	if err != nil {
		common.Log("promo-sign-reward-insert", "add records to query: %s err: %s", query, err.Error())
	}

	return
}

//signClearCache 清除会员签到记录
func signClearCache(startAt int64, username []string) {

	pipe := cli.TxPipeline()
	defer pipe.Close()

	for _, v := range username {
		key := fmt.Sprintf("%s:ps:%d", v, startAt)
		pipe.Unlink(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		common.Log("sign-clear-cache", "err: %s", err.Error())
	}
}

//signEsMemberTotal 获取时间区间中签到会员去重总数
func signEsMemberTotal(startAt, endAt int64) (int64, string, error) {

	query := elastic.NewBoolQuery().Must(
		elastic.NewRangeQuery("day").Gte(startAt).Lte(endAt),
	)

	agg := elastic.NewFiltersAggregation().Filter(elastic.NewMatchAllQuery()).OtherBucket(false).
		SubAggregation("username", elastic.NewCardinalityAggregation().Field("username"))
	resOrder, err := es.Search().Index(esPrefix+"tbl_promo_sign_record").
		Query(query).Size(0).Aggregation("group", agg).Do(ctx)
	if err != nil {
		return 0, "es", err
	}

	terms, ok := resOrder.Aggregations.Terms("group")
	if !ok {
		return 0, "", nil
	}

	var total int64 = 0
	if len(terms.Buckets) > 0 {

		n, ok := terms.Buckets[0].Cardinality("username")
		if ok {
			total = int64(*n.Value)
		}
	}

	return total, "", nil
}

//signMembers 分页读取签到会员帐号
func signMembers(startAt, endAt int64, page int64) ([]string, []interface{}, string, error) {

	query := elastic.NewBoolQuery().Must(
		elastic.NewRangeQuery("day").Gte(startAt).Lte(endAt),
	)
	offset := page * SignAwardPageSize
	collapseData := elastic.NewCollapseBuilder("username")
	fsc := elastic.NewFetchSourceContext(true).Include("username")
	resOrder, err := es.Search().Index(esPrefix + "tbl_promo_sign_record").FetchSourceContext(fsc).
		Collapse(collapseData).From(int(offset)).Size(SignAwardPageSize).Query(query).Do(ctx)
	if err != nil {
		return nil, nil, "es", err
	}

	var (
		users  []string
		unames []interface{}
		p      fastjson.Parser
	)

	for _, v := range resOrder.Hits.Hits {

		b, err := v.Source.MarshalJSON()
		if err != nil {
			return users, unames, "", err
		}

		buckets, err := p.ParseBytes(b)
		if err != nil {
			return users, unames, "", err
		}

		users = append(users, string(buckets.GetStringBytes("username")))
		unames = append(unames, string(buckets.GetStringBytes("username")))
	}

	return users, unames, "", nil
}

//signedList 获取redis签到记录
func signedList(startAt int64, users []string, conf signConfig) (map[string]map[string]string, string, error) {

	data := map[string]map[string]string{}
	pipe := cli.TxPipeline()
	defer pipe.Close()

	cmds := map[string]*redis.ZSliceCmd{}
	for _, v := range users {
		key := fmt.Sprintf("%s:ps:%d", v, startAt)
		cmd := pipe.ZRangeWithScores(ctx, key, 0, -1)
		cmds[v] = cmd
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return data, "redis", err
	}

	for k, v := range cmds {

		if v.Err() != nil {
			continue
		}

		res, err := v.Result()
		if err != nil {
			continue
		}

		n := len(res)
		if n == 0 {
			continue
		}

		if n > 7 {
			n = 7
		}

		resignAt := "0"
		amount := "0"
		waterFlow := "0"
		state := fmt.Sprintf("%d", common.SignPrizeUnReach)
		signDetail := []string{"0", "0", "0", "0", "0", "0", "0"}
		for rk, z := range res {

			if rk > 6 {
				continue
			}

			weekDay := int(z.Score)
			if weekDay > 7 {
				continue
			}

			signDetail[weekDay-1] = "1"

			//每周只能签到1次，则一但补签时间不为0时进入下一次循环
			if resignAt != "0" {
				continue
			}

			record := SignRecord{}
			err = helper.JsonUnmarshal([]byte(z.Member.(string)), &record)
			if err != nil {
				continue
			}

			if record.IsResign == 1 {
				resignAt = fmt.Sprintf("%d", record.SignAt)
			}
		}

		if n >= SignDefaultNum {

			cn := len(conf.Award)
			if cn > 0 {

				rn := len(conf.Award[cn-1].Rules)
				if rn > 0 {
					amount = conf.Award[cn-1].Rules[SignDefaultAward]
					waterFlow = conf.Award[cn-1].WaterFlow
					state = fmt.Sprintf("%d", common.SignPrizeWaitHandOut)
				}
			}
		}

		data[k] = map[string]string{
			"sign_day_num":     fmt.Sprintf("%d", n),
			"resign_at":        resignAt,
			"sign_detail":      strings.Join(signDetail, ","),
			"amount":           amount,
			"hand_out_state":   state,
			"water_flow":       waterFlow,
			"award_way":        fmt.Sprintf("%d", conf.AwardWay),
			"deposit_amount":   "0",
			"valid_bet_amount": "0",
			"net_win_amount":   "0",
		}
	}

	return data, "", nil
}

//signLoadConf 加载奖励配置
func signLoadConf() (signConfig, string, error) {

	data := signConfig{}
	val, err := cli.Get(ctx, "P:C:44").Result()
	if err != nil {
		return data, "redis", err
	}

	err = helper.JsonUnmarshal([]byte(val), &data)
	if err != nil {
		return data, "", err
	}

	return data, "", nil
}

//signRewardRecord 构建单会员奖金
func signRewardRecord(startAt, endAt int64, username string, data map[string]string, member common.Member) (g.Record, error) {

	signDayNum, ok := data["sign_day_num"]
	if !ok {
		return nil, nil
	}

	resignAt, ok := data["resign_at"]
	if !ok {
		return nil, nil
	}

	signDetail, ok := data["sign_detail"]
	if !ok {
		return nil, nil
	}

	depositAmount, ok := data["deposit_amount"]
	if !ok {
		return nil, nil
	}

	validBetAmount, ok := data["valid_bet_amount"]
	if !ok {
		return nil, nil
	}

	netWinAmount, ok := data["net_win_amount"]
	if !ok {
		return nil, nil
	}

	amountStr, ok := data["amount"]
	if !ok {
		return nil, nil
	}

	handOutState, ok := data["hand_out_state"]
	if !ok {
		return nil, nil
	}

	awardWay, ok := data["award_way"]
	if !ok {
		return nil, nil
	}

	reviewState := common.PromoGiftReviewing
	if awardWay == SignAwardWayManual {
		reviewState = common.PromoGiftReviewPass
	}

	amount, _ := decimal.NewFromString(amountStr)
	if amount.Cmp(zero) > 0 {
		handOutState = fmt.Sprintf("%d", common.SignPrizeWaitHandOut)
	}

	now := time.Now()
	s := now.Unix()
	//构建签到奖励记录
	record := g.Record{
		"id":               helper.GenId(),
		"uid":              member.UID,
		"username":         username,
		"week_start_at":    startAt,
		"week_end_at":      endAt,
		"sign_day_num":     signDayNum,
		"resign_at":        resignAt,
		"sign_detail":      signDetail,
		"deposit_amount":   depositAmount,
		"valid_bet_amount": validBetAmount,
		"net_win_amount":   netWinAmount,
		"award_way":        awardWay,
		"amount":           amountStr,
		"resettle":         0,
		"dividend_id":      0,
		"hand_out_state":   handOutState,
		"hand_out_at":      s,
		"review_state":     reviewState,
		"review_at":        0,
		"created_at":       s,
		"review_uid":       0,
		"review_name":      "",
		"review_remark":    "",
		"updated_at":       0,
		"updated_uid":      0,
		"updated_name":     "",
		"prefix":           prefix,
	}

	return record, nil
}

//signMemberReward 构建会员列表奖金记录
func signMemberReward(startAt, endAt int64, users []string, userList []interface{}, conf signConfig) ([]g.Record, string, error) {

	signed, flg, err := signedList(startAt, users, conf)
	if err != nil {
		return nil, flg, err
	}

	flg, err = signMemberBetRecord(startAt, endAt, userList, conf, signed)
	if err != nil {
		return nil, flg, err
	}

	members, flg, err := common.MemberMCache(db, users)
	if err != nil {
		return nil, flg, err
	}

	var result []g.Record
	for _, v := range users {

		mem, ok := members[v]
		if !ok {
			continue
		}

		sign, ok := signed[v]
		if !ok {
			continue
		}

		record, err := signRewardRecord(startAt, endAt, v, sign, mem)
		if err != nil {
			continue
		}

		if len(record) > 0 {
			result = append(result, record)
		}
	}

	return result, "", nil
}

//signMemberBetRecord 拉取时间区间中会员统计数据
func signMemberBetRecord(startAt, endAt int64, users []interface{}, conf signConfig, signed map[string]map[string]string) (string, error) {

	// 获取存款金额 提款金额 总输赢 有效投注额
	aggField := map[string]string{
		"company_net_amount": "company_net_amount",
		"deposit_amount":     "deposit_amount",
		"valid_bet_amount":   "valid_bet_amount",
	}

	agg := elastic.NewTermsAggregation().Field("username").Size(len(users))
	for name, field := range aggField {
		agg.SubAggregation(name, elastic.NewSumAggregation().Field(field))
	}

	query := elastic.NewBoolQuery().Filter(
		elastic.NewTermsQuery("username", users...),
		elastic.NewRangeQuery("report_time").Gte(startAt).Lte(endAt))

	search, err := reportEs.Search().Index(pullPrefix+"t_mem_settle_finance_report").Query(query).Size(0).Aggregation("username", agg).Do(ctx)
	if err != nil {
		return "es", err
	}

	aggRes, ok := search.Aggregations.Terms("username")
	if !ok {
		return "", nil
	}

	for _, v := range aggRes.Buckets {

		userName := v.Key.(string)

		uSign, ok := signed[userName]
		if !ok {
			uSign = map[string]string{
				"sign_day_num":     "0",
				"resign_at":        "0",
				"sign_detail":      "0,0,0,0,0,0,0",
				"amount":           "0",
				"hand_out_state":   fmt.Sprintf("%d", common.SignPrizeUnReach),
				"deposit_amount":   "0",
				"valid_bet_amount": "0",
				"net_win_amount":   "0",
				"water_flow":       "0",
				"award_way":        fmt.Sprintf("%d", conf.AwardWay),
			}
		}

		amount, _ := v.Sum("deposit_amount")
		uSign["deposit_amount"] = fmt.Sprintf("%0.4f", *amount.Value)
		amount, _ = v.Sum("company_net_amount")
		uSign["net_win_amount"] = fmt.Sprintf("%0.4f", (*amount.Value)*(-1))
		amount, _ = v.Sum("valid_bet_amount")
		uSign["valid_bet_amount"] = fmt.Sprintf("%0.4f", *amount.Value)

		uAmount := decimal.NewFromFloat(*amount.Value)
		for _, v := range conf.Award {

			waterFlow, err := decimal.NewFromString(v.WaterFlow)
			if err != nil {
				common.Log("sign-member-bet-record", "sign-member-bet-record err: %s", err.Error())
				continue
			}

			if uAmount.Cmp(waterFlow) >= 0 {

				signNum, ok := uSign["sign_day_num"]
				if !ok {
					signNum = "0"
				}

				num, _ := strconv.Atoi(signNum)
				rn := len(v.Rules)
				if num > rn {
					num = rn
				}

				if num > 0 {
					uSign["amount"] = v.Rules[num-1]
					uSign["water_flow"] = v.WaterFlow
				}

				break
			}
		}

		signed[userName] = uSign
	}

	return "", nil
}

func (i SignRecord) MarshalBinary() ([]byte, error) {
	return helper.JsonMarshal(i)
}

//SignLoadRecord 加载上一周签到记录
func SignLoadRecord(day int64) {

	lastWeekStartAt := helper.WeekTST(day, loc).AddDate(0, 0, -7).Unix()
	lastWeekEndAt := helper.WeekTET(day, loc).AddDate(0, 0, -7).Unix()

	total, flg, err := signEsMemberTotal(lastWeekStartAt, lastWeekEndAt)
	if err != nil {
		common.Log("promo-sign-total", "es members_total %s err: %s", flg, err.Error())
		return
	}

	pages := total / SignAwardPageSize
	if total%SignAwardPageSize != 0 {
		pages = pages + 1
	}

	var page int64
	for ; page < pages; page++ {

		query := elastic.NewBoolQuery().Must(
			elastic.NewRangeQuery("day").Gte(lastWeekStartAt).Lte(lastWeekEndAt),
		)

		offset := page * SignAwardPageSize
		fields := []string{"uid", "username", "deposit_amount", "valid_bet_amount", "is_resign", "device", "sign_at", "day"}
		collapseData := elastic.NewCollapseBuilder("username")
		collapseData.InnerHit(elastic.NewInnerHit().Name("username").Sort("day", true).Size(10))
		fsc := elastic.NewFetchSourceContext(true).Include(fields...)
		resOrder, err := es.Search().Index(esPrefix + "tbl_promo_sign_record").FetchSourceContext(fsc).
			Collapse(collapseData).From(int(offset)).Size(SignAwardPageSize).Query(query).Do(ctx)
		if err != nil {
			return
		}

		data := map[string][]SignRecord{}
		for _, v := range resOrder.Hits.Hits {

			username := ""
			var item []SignRecord
			for _, vv := range v.InnerHits["username"].Hits.Hits {

				record := SignRecord{}
				err := helper.JsonUnmarshal(vv.Source, &record)
				if err != nil {
					common.Log("load-record-sign", "err: %s", err.Error())
					continue
				}

				record.ID = vv.Id
				if len(username) == 0 {
					username = record.Username
				}

				item = append(item, record)
			}

			data[username] = item
		}

		pipe := cli.TxPipeline()
		for k, v := range data {
			key := fmt.Sprintf("%s:ps:%d", k, lastWeekStartAt)
			days := map[int]interface{}{}
			for _, vv := range v {

				t := time.Unix(vv.Day, 0).In(loc)
				weekDay := int(t.Weekday())
				if weekDay == 0 {
					weekDay = 7
				}

				_, ok := days[weekDay]
				if ok {
					continue
				}

				days[weekDay] = weekDay
				data := redis.Z{
					Score:  float64(weekDay),
					Member: vv,
				}

				pipe.ZAdd(ctx, key, &data)
			}
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			common.Log("load-record-sign", "err: %s", err.Error())
		}

		_ = pipe.Close()
	}
}
