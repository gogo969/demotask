package promo

import (
	"fmt"
	g "github.com/doug-martin/goqu/v9"
	"github.com/valyala/fastjson"
	"math"
	"strings"
	"task/contrib/helper"
	"task/modules/common"
	"time"
)

// promoStateChange 活动自动启停
func promoStateChange(m map[string]interface{}) {

	state, ok := m["state"]
	if !ok {
		return
	}

	id, ok := m["id"]
	if !ok {
		return
	}

	ex := g.Ex{
		"id": id.(string),
	}
	record := g.Record{
		"state": state.(string),
	}

	var pd promoData
	query, _, _ := dialect.From("tbl_promo").Select(field...).Where(ex).ToSQL()
	err := db.Get(&pd, query)
	if err != nil {
		common.Log("promo", "error: %v", err)
		return
	}

	t := 0
	now := time.Now().Unix()
	if pd.StartAt < pd.DisplayAt {
		ts := math.Abs(float64(pd.StartAt - now))
		t = int(ts)
	} else {
		ts := math.Abs(float64(pd.DisplayAt - now))
		t = int(ts)
	}

	//启用
	if m["state"] == "1" && t > 10 {
		common.Log("promo", "time state too large, drop")
		return
	}

	if pd.Period == 0 {

		if pd.HideAt < pd.EndAt {
			t = int(math.Abs(float64(pd.EndAt - now)))
			common.Log("promo", "end time state too large, drop")
		} else {
			t = int(math.Abs(float64(pd.HideAt - now)))
			common.Log("promo", "hide time state too large, drop")
		}

		//停用
		if m["state"] == "0" && pd.Period == 0 && t > 10 {
			common.Log("promo", "time state too large, drop")
			return
		}
	}

	query, _, _ = dialect.Update("tbl_promo").Set(record).Where(ex).ToSQL()
	_, err = db.Exec(query)
	if err != nil {
		common.Log("promo", "update error: %s, sql: %s", err.Error(), query)
	}

	_ = promosToCache()
}

func promosToCache() error {

	var (
		data []Promo
		ids  []string
	)
	ex := g.Ex{"state": 1, "prefix": prefix}
	query, _, _ := dialect.From("tbl_promo").Select(colPromo...).Where(ex).Order(g.C("sort").Asc()).ToSQL()
	err := db.Select(&data, query)
	if err != nil {
		return err
	}

	pipe := cli.TxPipeline()
	defer pipe.Close()

	listData := map[string][]map[string]interface{}{}
	promoOnline := map[string]interface{}{}
	var hover []map[string]string
	for _, v := range data {

		t := 0
		if v.Period == 0 {
			now := time.Now().Unix()
			if v.HideAt < v.EndAt {
				t = int(math.Abs(float64(v.EndAt - now)))
			} else {
				t = int(math.Abs(float64(v.HideAt - now)))
			}

			if v.Period == 0 && t < 0 {
				//记录活动已过期但未变更状态的活动id
				ids = append(ids, v.ID)
				continue
			}
		}

		p := map[string]interface{}{
			"id":       v.ID,
			"start_at": v.StartAt,
			"state":    v.State,
			"end_at":   v.EndAt,
			"period":   v.Period,
			"title":    v.Title,
		}

		pd, _ := helper.JsonMarshal(p)
		promoOnline[v.ID] = string(pd)

		item := map[string]interface{}{
			"id":           v.ID,
			"title":        v.Title,
			"flags":        v.Flags,
			"tag":          v.Tag,
			"ty":           v.Ty,
			"content_ty":   v.ContentTy,
			"sort":         v.Sort,
			"period":       v.Period,
			"display_time": v.DisplayAt,
			"hide_time":    v.HideAt,
			"start_at":     v.StartAt,
			"end_at":       v.EndAt,
			"start_time":   parseTime(v.StartAt),
			"end_time":     parseTime(v.EndAt),
			"images":       jsonToMap(v.Images),
			"h5_images":    jsonToMap(v.H5Images),
		}

		for _, dv := range strings.Split(v.Devices, ",") {
			for _, iv := range strings.Split(devices[dv], ",") {

				key := fmt.Sprintf("PM:%d:%s", v.Tag, iv)
				listData[key] = append(listData[key], item)
				key = fmt.Sprintf("PM:%d:%s", 0, iv)
				listData[key] = append(listData[key], item)
			}
		}

		if v.FloatingIcon == PromoFloatingOpen {

			hItem := map[string]string{
				"pid": v.ID,
				"ty":  fmt.Sprintf("%d", v.Ty),
			}

			hover = append(hover, hItem)
		}
	}

	for k, v := range listData {

		body, _ := helper.JsonMarshal(v)
		fmt.Printf("key=%s,value=%s\n", k, body)
		pipe.Unlink(ctx, k)
		pipe.Set(ctx, k, string(body), 100*time.Hour)
		pipe.Persist(ctx, k)
	}

	hdata := map[string]interface{}{
		"promo": hover,
	}

	key := "P:H"
	pipe.Unlink(ctx, key)
	body, _ := helper.JsonMarshal(hdata)
	pipe.Set(ctx, key, string(body), 100*time.Hour)
	pipe.Persist(ctx, key)

	//将所有打开的活动写入到redis中
	key = "P:O"
	pipe.Unlink(ctx, key)
	pipe.HMSet(ctx, key, promoOnline)
	pipe.Persist(ctx, key)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	if len(ids) > 0 {
		//修改即将下线的活动
		ex = g.Ex{"id": ids}
		query, _, _ = dialect.Update("tbl_promo").Set(g.Record{"state": 0}).Where(ex).ToSQL()
		_, err = db.Exec(query)
		if err != nil {
			common.Log("promo", "close promo error: %s, sql: %s", err.Error(), query)
		}
	}

	return nil
}

func parseTime(t int64) string {

	if t == 0 {
		return ""
	}

	ts := time.Unix(t, 0)
	return ts.Format("02/01/2006")
}

func jsonToMap(val string) map[string]string {

	if len(val) == 0 {
		return nil
	}

	var p fastjson.Parser
	data := map[string]string{}
	images, err := p.Parse(val)
	if err != nil {
		return nil
	}

	items := images.GetArray("images")
	if len(items) > 0 {
		data["images"] = string(items[0].GetStringBytes())
	}

	items = images.GetArray("list")
	if len(items) > 0 {
		data["list"] = string(items[0].GetStringBytes())
	}

	items = images.GetArray("back_ground")
	if len(items) > 0 {
		data["back_ground"] = string(items[0].GetStringBytes())
	}

	items = images.GetArray("horizontal")
	if len(items) > 0 {
		data["horizontal"] = string(items[0].GetStringBytes())
	}

	items = images.GetArray("vertical")
	if len(items) > 0 {
		data["vertical"] = string(items[0].GetStringBytes())
	}

	items = images.GetArray("share")
	if len(items) > 0 {
		data["share"] = string(items[0].GetStringBytes())
	}

	return data
}
