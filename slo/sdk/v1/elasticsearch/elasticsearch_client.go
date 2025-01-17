package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	sloapi "github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/manager/handler"
	es7 "github.com/olivere/elastic/v7"
)

var _ handler.RecordStorage = &ElasticsearchAPI{}

type ElasticsearchAPI struct {
	*es7.Client
	Suffix string
}

func NewElasticElasticsearchAPIsearchAPI(client *es7.Client, esSuffix string) *ElasticsearchAPI {
	return &ElasticsearchAPI{Client: client, Suffix: esSuffix}
}

func NewElasticSearchAPI(cfg *ElasticsearchConfig) (*ElasticsearchAPI, error) {
	if cfg == nil {
		return nil, nil
	}

	errorLog := log.New(os.Stdout, "esclient", log.LstdFlags)
	esClient, err := es7.NewClient(
		es7.SetErrorLog(errorLog),
		es7.SetURL(cfg.URLS...),
		es7.SetSniff(false),
		es7.SetBasicAuth(cfg.Username, cfg.Password),
	)

	if err != nil {
		return nil, err
	}

	return &ElasticsearchAPI{
		Client: esClient,
		Suffix: cfg.EsIndexSuffix,
	}, nil
}

func (esapi *ElasticsearchAPI) StoreSLOResult(results []*model.SLOResult, startTSMillis int64, step time.Duration) {
	index, _ := SLORecordESPattern.GetIndexWithTimePattern(startTSMillis)
	indexTimestamp := time.Now()
	bulkService := esapi.Bulk()

	var checkedEntry int = 0
	for i := 0; i < len(results); i++ {
		for s := 0; s < len(results[i].SLOGroup); s++ {
			if results[i].SLOGroup[s].Status == model.Unknown {
				continue
			}
			record := &model.SLOResultRecord{
				SLOServiceName: &results[i].SLOServiceName,
				SLOGroup:       &results[i].SLOGroup[s],
				Step:           model.GetRecordStepFromDuration(step),
				IndexTimestamp: indexTimestamp.Unix(),
			}
			checkedEntry++

			bulkService.Add(es7.NewBulkIndexRequest().Index(index).Doc(record))
		}
	}

	if checkedEntry == 0 {
		return
	}

	response, err := bulkService.Do(context.Background())
	if response != nil {
		failed := response.Failed()
		if len(failed) > 0 {
			log.Printf("store sloResult failed: Indexed %d documents (failures=%d)\n", len(response.Items), len(failed))
		}
	}

	if err != nil {
		log.Printf("store sloResult error: %v", err)
		return
	}
}

func (esapi *ElasticsearchAPI) SearchSLOResult(
	entryURL string,
	startMS int64,
	endMS int64,
	pageParam *api.PageParam,
	duration time.Duration,
	skipInactiveEntry bool,
	skipHealthyEntry bool,
	options ...api.SortByOption,
) (result []*model.SLOResult, count int, err error) {
	preResults, err := esapi.searchSLOResult(entryURL, startMS, endMS, duration, skipHealthyEntry, options...)
	if err != nil {
		return []*model.SLOResult{}, 0, err
	}

	from, to := pageByGolang(len(preResults), pageParam)
	result, err = esapi.fillSLOGroup(preResults[from:to], startMS, endMS, duration)
	if err != nil {
		return []*model.SLOResult{}, 0, err
	}

	return result, len(preResults), nil
}

func (esapi *ElasticsearchAPI) searchSLOResult(entryUri string, startTime int64, endTime int64, duration time.Duration, skipHealthyEntry bool, options ...sloapi.SortByOption) (results []*model.SLOResult, err error) {
	query := es7.NewBoolQuery().Must(
		es7.NewRangeQuery("sloGroup.endTime").Gt(startTime).Lte(endTime),
		es7.NewTermQuery("step", model.GetRecordStepFromDuration(duration)),
	)

	if len(entryUri) > 0 {
		query.Must(es7.NewWildcardQuery("serviceName.entryUri", "*"+entryUri+"*"))
	}

	groupByUrls := es7.NewTermsAggregation().Field("serviceName.entryUri").Size(1000).
		SubAggregation(sloapi.SortByNotAchievedCount, es7.NewFilterAggregation().Filter(
			es7.NewBoolQuery().Must(es7.NewTermQuery("sloGroup.status", "NotAchieved")),
		)).
		SubAggregation(sloapi.SortByRequestCount, es7.NewSumAggregation().Field("sloGroup.requestCount"))

	for _, option := range options {
		groupByUrls.Order(option, false)
	}

	res, err := esapi.
		Search("slo_result*").
		Query(query).
		Sort("sloGroup.startTime", true).
		Aggregation("groupByUrls", groupByUrls).
		Size(0).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	urlGroups, success := res.Aggregations.Terms("groupByUrls")
	if !success {
		return nil, nil
	}

	for i := 0; i < len(urlGroups.Buckets); i++ {
		urlBucket := urlGroups.Buckets[i]
		if skipHealthyEntry {
			notAchievedCount, find := urlBucket.Aggregations.Filter(sloapi.SortByNotAchievedCount)
			if find && notAchievedCount.DocCount < 1 {
				continue
			}
		}
		results = append(results, &model.SLOResult{
			SLOServiceName: model.SLOServiceName{
				EntryUri: urlBucket.Key.(string),
			},
			SLOGroup: []model.SLOGroup{},
		})
	}

	return results, nil
}

func (esapi *ElasticsearchAPI) fillSLOGroup(results []*model.SLOResult, startTime int64, endTime int64, duration time.Duration) ([]*model.SLOResult, error) {
	query := es7.NewBoolQuery().Must(
		es7.NewRangeQuery("sloGroup.endTime").Gt(startTime).Lte(endTime),
		es7.NewTermQuery("step", model.GetRecordStepFromDuration(duration)),
	)

	for _, target := range results {
		query.Should(es7.NewTermQuery("serviceName.entryUri", target.SLOServiceName.EntryUri))
	}

	entryUrlCollapse := es7.NewCollapseBuilder("serviceName.entryUri").
		InnerHit(es7.NewInnerHit().
			Name("records").
			Size(100).
			Sort("sloGroup.startTime", true),
		)

	res, err := esapi.
		Search("slo_result*").
		Query(query).
		Collapse(entryUrlCollapse).
		Size(100).
		Do(context.Background())

	if err != nil {
		return nil, err
	}

	for _, hit := range res.Hits.Hits {
		contentKey, find1 := parseStringFields(hit.Fields, "serviceName.entryUri", 0)
		records, find2 := hit.InnerHits["records"]
		if !find1 || !find2 {
			continue
		}

		var result *model.SLOResult
		for i := 0; i < len(results); i++ {
			if results[i].SLOServiceName.EntryUri == contentKey {
				result = results[i]
			}
		}

		if result == nil {
			continue
		}

		for _, recordJson := range records.Hits.Hits {
			record := model.SLOResultRecord{}
			err := json.Unmarshal(recordJson.Source, &record)
			if err != nil {
				return nil, err
			}

			result.SLOGroup = append(result.SLOGroup, *record.SLOGroup)
		}
	}
	return results, nil
}

func (esapi *ElasticsearchAPI) QueryTimeSeriesRootCauseCount(index string, entryUri *string, startTime int64, endTime int64, step time.Duration) (model.RootCauseCountTimeSeries, error) {
	if startTime >= endTime {
		return []model.RootCauseCountPoint{}, fmt.Errorf("query range is negative, %d ~ %d", startTime, endTime)
	}

	if entryUri == nil || len(*entryUri) == 0 {
		return []model.RootCauseCountPoint{}, fmt.Errorf("entryKey must no be null")
	}

	var points model.RootCauseCountTimeSeries
	var err error
	switch step {
	case time.Minute:
		points, err = esapi.queryRangeCausePerMin(index, *entryUri, startTime, endTime, step)
	default:
		points, err = esapi.queryRangeCause(index, *entryUri, startTime, endTime, step)
	}
	return points, err
}

func (esapi *ElasticsearchAPI) queryRangeCause(
	index string,
	entryUri string,
	startTime int64,
	endTime int64,
	step time.Duration,
) (model.RootCauseCountTimeSeries, error) {
	query := es7.NewBoolQuery().
		Filter(es7.NewTermQuery("data.content_key", entryUri)).
		Filter(es7.NewTermQuery("is_drop", "false")).
		Filter(es7.NewRangeQuery("data.end_time").Gte(startTime).Lte(endTime))

		// Aggregations
	// {"aggregations":{"timeBucket":{"aggregations":{"rootCause":{"terms":{"field":"data.mutated_service"}}},"histogram":{"field":"data.end_time","interval":15000000000,"offset":0}}}}
	rootCauseAgg := es7.NewTermsAggregation().Field("data.mutated_service")
	timeBucketAgg := es7.NewHistogramAggregation().
		Field("data.end_time").
		Interval(float64(step)).
		MinDocCount(0).
		ExtendedBounds(
			float64(startTime),
			float64(endTime),
		).
		Offset(float64(startTime%int64(step))).
		SubAggregation("rootCause", rootCauseAgg)

	result, err := esapi.
		Search(index+"_"+esapi.Suffix).
		Query(query).
		Aggregation("timeBucket", timeBucketAgg).
		IgnoreUnavailable(true).
		AllowNoIndices(true).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Printf("Not found root cause counts: %v", err)
		return []model.RootCauseCountPoint{}, err
	}

	timeBucket, ok := result.Aggregations.Histogram("timeBucket")
	if !ok {
		log.Printf("Not found record when querying cause counts for entry %s", entryUri)
		return []model.RootCauseCountPoint{}, nil
	}

	var points = make([]model.RootCauseCountPoint, 0, len(timeBucket.Buckets))
	for i := 0; i < len(timeBucket.Buckets); i++ {
		bucket := timeBucket.Buckets[i]
		// Nano timestamp to Milliseconds
		timestamp := int64(bucket.Key / 1e6)
		termsByRootCause, ok := bucket.Terms("rootCause")
		if !ok {
			points = append(points, model.RootCauseCountPoint{
				Timestamp:         timestamp,
				RootCauseCountMap: make(map[string]int),
			})
			continue
		}

		causeCountMap := make(map[string]int)
		for _, rootCauseBucket := range termsByRootCause.Buckets {
			key := rootCauseBucket.Key.(string)
			causeCountMap[key] = int(rootCauseBucket.DocCount)
		}

		points = append(points, model.RootCauseCountPoint{
			Timestamp:         timestamp,
			RootCauseCountMap: causeCountMap,
		})
	}
	return points, nil
}

func (esapi *ElasticsearchAPI) queryRangeCausePerMin(
	index string,
	entryUri string,
	startTime int64,
	endTime int64,
	step time.Duration,
) (model.RootCauseCountTimeSeries, error) {
	bucketDuration := step / 4

	query := es7.NewBoolQuery().
		Filter(es7.NewTermQuery("data.content_key", entryUri)).
		Filter(es7.NewTermQuery("is_drop", "false")).
		Filter(es7.NewRangeQuery("data.end_time").Gte(startTime - bucketDuration.Nanoseconds()).Lte(endTime))

	// Aggregations
	// {"aggregations":{"timeBucket":{"aggregations":{"rootCause":{"terms":{"field":"data.mutated_service"}}},"histogram":{"field":"data.end_time","interval":15000000000,"offset":0}}}}
	rootCauseAgg := es7.NewTermsAggregation().Field("data.mutated_service")
	timeBucketAgg := es7.NewHistogramAggregation().
		Field("data.end_time").
		Interval(float64(bucketDuration)).
		Offset(float64(startTime%int64(bucketDuration))).
		MinDocCount(0).
		ExtendedBounds(
			float64(startTime-bucketDuration.Nanoseconds()),
			float64(endTime-bucketDuration.Nanoseconds()),
		).
		SubAggregation("rootCause", rootCauseAgg)

	result, err := esapi.
		Search(index+"_"+esapi.Suffix).
		Query(query).
		Aggregation("timeBucket", timeBucketAgg).
		IgnoreUnavailable(true).
		AllowNoIndices(true).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Printf("Not found root cause counts: %v", err)
		return []model.RootCauseCountPoint{}, err
	}

	timeBucket, ok := result.Aggregations.Histogram("timeBucket")
	if !ok {
		log.Printf("Not found record when querying cause counts for entry %s", entryUri)
		return []model.RootCauseCountPoint{}, nil
	}

	var points = make([]model.RootCauseCountPoint, 0, len(timeBucket.Buckets)/4+1)
	// Bucket to Points
	// 0,1,2,3,4 -> point1
	// 4,5,6,7,8 -> point2
	// 8,9,10,11,12 -> point3
	// ...
	for i := 0; i+4 < len(timeBucket.Buckets); i = i + 4 {
		// Nano timestamp to Milliseconds
		timestamp := int64(timeBucket.Buckets[i+1].Key / 1e6)
		point, err := sumBucketsToPoints(timeBucket.Buckets[i:i+5], timestamp)
		if err != nil {
			continue
		}
		points = append(points, point)
	}
	return points, nil
}

func sumBucketsToPoints(buckets []*es7.AggregationBucketHistogramItem, timestamp int64) (model.RootCauseCountPoint, error) {
	var res = model.RootCauseCountPoint{
		Timestamp:         timestamp,
		RootCauseCountMap: map[string]int{},
	}

	for _, bucket := range buckets {
		termsByRootCause, ok := bucket.Terms("rootCause")
		if !ok {
			continue
		}

		for _, rootCauseBucket := range termsByRootCause.Buckets {
			key := rootCauseBucket.Key.(string)
			value := int(rootCauseBucket.DocCount)
			if oldValue, find := res.RootCauseCountMap[key]; find {
				res.RootCauseCountMap[key] = oldValue + value
			} else {
				res.RootCauseCountMap[key] = value
			}
		}
	}

	return res, nil
}

func parseStringFields(fields es7.SearchHitFields, fieldName string, index int) (string, bool) {
	fieldsData, find := fields[fieldName]
	if !find {
		return "", false
	}

	switch fieldsData := fieldsData.(type) {
	case string:
		return fieldsData, true
	case []interface{}:
		if len(fieldsData) > index {
			return fieldsData[index].(string), true
		}
	}
	return "", false
}

func pageByGolang(count int, pageParam *api.PageParam) (from int, to int) {
	if pageParam == nil {
		return 0, count
	}
	var endTo = pageParam.PageNum * pageParam.PageSize
	if endTo > count {
		endTo = count
	}
	return (pageParam.PageNum - 1) * pageParam.PageSize, endTo
}
