package common

import (
	"fmt"
	"github.com/olivere/elastic/v7"
)

func EsQuerySearch(esCli *elastic.Client, index, sortField string, page, pageSize int,
	fields []string, boolQuery *elastic.BoolQuery, agg map[string]*elastic.SumAggregation) (int64, []*elastic.SearchHit, elastic.Aggregations, string, error) {

	fsc := elastic.NewFetchSourceContext(true).Include(fields...)
	offset := (page - 1) * pageSize
	//打印es查询json
	esService := esCli.Search().FetchSourceContext(fsc).Query(boolQuery).From(offset).Size(pageSize).TrackTotalHits(true).Sort(sortField, false)
	for k, v := range agg {
		esService = esService.Aggregation(k, v)
	}
	resOrder, err := esService.Index(index).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return 0, nil, nil, "es", err
	}

	if resOrder.Status != 0 || resOrder.Hits.TotalHits.Value <= int64(offset) {
		return resOrder.Hits.TotalHits.Value, nil, nil, "", nil
	}

	return resOrder.Hits.TotalHits.Value, resOrder.Hits.Hits, resOrder.Aggregations, "", nil
}
