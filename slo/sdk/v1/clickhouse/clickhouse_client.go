package clickhouse

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/manager/handler"
)

var _ handler.RecordStorage = &ClickhouseAPI{}

type ClickhouseAPI struct {
	cfg  *ClickhouseConfig
	conn driver.Conn

	fullTableName  string
	fieldNamesList []string
	fieldNames     string
	Parts          []PartFromSLOResult
}

func NewClickhouseAPI(cfg *ClickhouseConfig) (*ClickhouseAPI, error) {
	if cfg == nil {
		return nil, nil
	}
	conn, err := newConn(cfg)
	if err != nil {
		return nil, err
	}
	var record model.SLOResultRecord
	fieldNames, _ := buildTableFieldsDDL(reflect.TypeOf(record))

	var parts []PartFromSLOResult
	var validFieldName []string
	for _, name := range fieldNames {
		f := getPartFromSLOResult(name)
		if f == nil {
			log.Printf("field '%s' is not defined in PartFromSLOResult, ignore this field when insert", name)
		}
		parts = append(parts, f)
		validFieldName = append(validFieldName, name)
	}
	return &ClickhouseAPI{
		cfg:            cfg,
		conn:           conn,
		fullTableName:  fmt.Sprintf("`%s`.`%s`", cfg.Authentication.PlainText.Database, cfg.Table),
		fieldNames:     strings.Join(validFieldName, ","),
		fieldNamesList: validFieldName,
		Parts:          parts,
	}, nil
}

func (c *ClickhouseAPI) StoreSLOResult(results []*model.SLOResult, startTSMillis int64, step time.Duration) {
	indexTimestamp := time.Now()
	stepStr := model.GetRecordStepFromDuration(step)

	var valueLines = 0
	var valueBuffer strings.Builder
	for i := 0; i < len(results); i++ {
		if valueLines > 0 {
			valueBuffer.WriteByte(',')
		}
		valueLines += c.expandToClickhouseValue(&valueBuffer, results[i], string(stepStr), indexTimestamp.UnixMilli())
		if valueLines > 100 {
			insertSql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", c.fullTableName, c.fieldNames, valueBuffer.String())
			err := c.conn.AsyncInsert(context.Background(), insertSql, false)
			if err != nil {
				log.Printf("store sloResult error: %v", err)
			}
			valueLines = 0
		}
	}

	if valueLines > 0 {
		insertSql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", c.fullTableName, c.fieldNames, valueBuffer.String())
		err := c.conn.AsyncInsert(context.Background(), insertSql, false)
		if err != nil {
			log.Printf("store sloResult error: %v", err)
		}
	}
}

func (c *ClickhouseAPI) SearchSLOResult(
	entryURL string,
	startMS int64,
	endMS int64,
	pageParam *api.PageParam,
	duration time.Duration,
	skipInactiveEntry bool,
	skipHealthyEntry bool,
	options ...api.SortByOption,
) (result []*model.SLOResult, count int, err error) {
	filters := extractFilters(startMS, endMS, duration, skipHealthyEntry, entryURL)
	orders := extractOrders(options)

	entryCount, err := c.entryCount(filters)
	if entryCount == 0 || err != nil {
		return nil, entryCount, err
	}

	withEntry := c.preparePagedEntry(filters, orders, pageParam)
	recordRequest := `WITH pagedEntry AS ( %s )
SELECT %s FROM %s
GLOBAL JOIN pagedEntry ON slo_record.entryUri = pagedEntry.entryUri
WHERE %s ORDER BY %s , endTime ASC LIMIT 5000`

	quest := fmt.Sprintf(
		recordRequest,
		withEntry,
		c.fieldNames,
		c.fullTableName,
		filters, orders)

	rows, err := c.conn.Query(context.Background(), quest)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	cachedSlo := map[string]*model.SLOResult{}
	for rows.Next() {
		var sloRecord SLORecordVO
		err := rows.ScanStruct(&sloRecord)
		if err != nil {
			continue
		}
		sloResult, exist := cachedSlo[sloRecord.EntryUri]
		if !exist {
			sloResult = initSLOResultFromRecord(sloRecord)
			result = append(result, sloResult)
			cachedSlo[sloRecord.EntryUri] = sloResult
		} else {
			addRecordToSLOResult(sloResult, sloRecord)
		}
	}

	return result, entryCount, nil
}

func initSLOResultFromRecord(sloRecord SLORecordVO) *model.SLOResult {
	sloResult := &model.SLOResult{
		SLOServiceName: model.SLOServiceName{
			EntryUri:     sloRecord.EntryUri,
			EntryService: "",
			Alias:        "",
		},
		SLOGroup: []model.SLOGroup{},
	}
	addRecordToSLOResult(sloResult, sloRecord)
	return sloResult
}

func addRecordToSLOResult(sloResult *model.SLOResult, sloRecord SLORecordVO) {
	group := model.SLOGroup{
		StartTime:           sloRecord.StartTime,
		EndTime:             sloRecord.EndTime,
		RequestCount:        int(sloRecord.RequestCount),
		Status:              model.SLOStatus(sloRecord.Status),
		SLOs:                []model.SLO{},
		SlowRootCauseCount:  map[string]int{},
		ErrorRootCauseCount: map[string]int{},
	}

	for index, sloType := range sloRecord.SLOsType {
		group.SLOs = append(group.SLOs, model.SLO{
			SLOConfig: &model.SLOConfig{
				Type:          model.SLOType(sloType),
				Multiple:      sloRecord.SLOsMultiple[index],
				ExpectedValue: sloRecord.SLOsExpectedValue[index],
				Source:        model.ExpectedSource(sloRecord.SLOsStatus[index]),
			},
			CurrentValue: sloRecord.SLOsCurrentValue[index],
			Status:       model.SLOStatus(sloRecord.SLOsStatus[index]),
		})
	}

	for index, slowService := range sloRecord.SlowRootCauseCountKey {
		group.SlowRootCauseCount[slowService] = int(sloRecord.SlowRootCauseCountValue[index])
	}
	for index, errorService := range sloRecord.ErrorRootCauseCountKey {
		group.ErrorRootCauseCount[errorService] = int(sloRecord.ErrorRootCauseCountValue[index])
	}

	sloResult.SLOGroup = append(sloResult.SLOGroup, group)
}

func extractOrders(options []api.SortByOption) string {
	var orderRule []string
	for _, option := range options {
		orderRule = append(orderRule, fmt.Sprintf("%s DESC", option))
	}
	orderRule = append(orderRule, "entryUri ASC")
	return strings.Join(orderRule, ", ")
}

func extractFilters(startMS int64, endMS int64, duration time.Duration, skipHealthyEntry bool, entryURL string) string {
	var filters []string
	filters = append(filters, fmt.Sprintf("endTime > %d AND endTime < %d", startMS, endMS))
	filters = append(filters, fmt.Sprintf("step = '%s'", model.GetRecordStepFromDuration(duration)))
	if skipHealthyEntry {
		filters = append(filters, fmt.Sprintf("%s > 0", api.SortByNotAchievedCount))
	}
	if len(entryURL) > 0 {
		entryURL = quotationMarksReplacer.Replace(entryURL)

		filters = append(filters, fmt.Sprintf("entryUri LIKE '%s%%'", entryURL))
	}
	return strings.Join(filters, " AND ")
}

var quotationMarksReplacer = strings.NewReplacer(`"`, `#`, `'`, `#`)

func (c *ClickhouseAPI) preparePagedEntry(filters string, order string, pageParam *api.PageParam) string {
	entryQuestTemplate := `SELECT entryUri,
	SUM( CASE WHEN status = 'NotAchieved' THEN 1 ELSE 0 END ) as %s,
	SUM( requestCount ) AS %s
FROM %s WHERE %s GROUP BY entryUri ORDER BY %s LIMIT %s`

	limit := "999 OFFSET 0"
	if pageParam != nil {
		limit = fmt.Sprintf("%d OFFSET %d", pageParam.PageSize, (pageParam.PageNum-1)*pageParam.PageSize)
	}

	return fmt.Sprintf(
		entryQuestTemplate,
		api.SortByNotAchievedCount, api.SortByRequestCount,
		c.fullTableName,
		filters,
		order,
		limit,
	)
}

func (c *ClickhouseAPI) entryCount(filters string) (int, error) {
	query := `SELECT COUNT(DISTINCT entryUri) FROM %s WHERE %s`
	var entryCount uint64 = 0
	err := c.conn.QueryRow(context.Background(), fmt.Sprintf(query, c.fullTableName, filters)).Scan(&entryCount)
	return int(entryCount), err
}

func (c *ClickhouseAPI) QueryTimeSeriesRootCauseCount(index string, entryUri *string, startTime int64, endTime int64, step time.Duration) (model.RootCauseCountTimeSeries, error) {
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
		points, err = c.queryRangeCausePerMin(index, *entryUri, startTime, endTime, step)
	default:
		points, err = c.queryRangeCause(index, *entryUri, startTime, endTime, step)
	}
	return points, err
}

func (c *ClickhouseAPI) queryRangeCause(
	tableName string,
	entryUri string,
	startTime int64,
	endTime int64,
	step time.Duration,
) (model.RootCauseCountTimeSeries, error) {
	bucketDuration := step.Nanoseconds()
	timeBucket := fmt.Sprintf("((end_time - %d) DIV %d)", startTime, bucketDuration)

	fieldBuilder := NewFieldBuilder().
		Alias("labels['mutated_service']", "mutated_service").
		Alias(timeBucket, "time_bucket").
		Alias("COUNT(1)", "total")
	whereClause := NewQueryBuilder().
		Between("end_time", uint64(startTime), uint64(endTime)).
		Equals("labels['content_key']", entryUri).
		Equals("is_drop", false)
	byBuilder := NewByLimitBuilder().
		GroupBy("mutated_service", "time_bucket")

	sql := fmt.Sprintf("SELECT %s FROM %s %s %s", fieldBuilder.String(), tableName, whereClause.String(), byBuilder.String())
	var results []struct {
		Service    string `ch:"mutated_service"`
		TimeBucket int64  `ch:"time_bucket"`
		Total      uint64 `ch:"total"`
	}
	if err := c.conn.Select(context.Background(), &results, sql, whereClause.values...); err != nil {
		log.Printf("Not found root cause counts: %v", err)
		return []model.RootCauseCountPoint{}, err
	}

	groups := map[int64]map[string]int{}
	for _, result := range results {
		timeBucket, exist := groups[result.TimeBucket]
		if !exist {
			timeBucket = map[string]int{}
		}
		timeBucket[result.Service] = int(result.Total)
		groups[result.TimeBucket] = timeBucket
	}

	var points = make([]model.RootCauseCountPoint, 0)
	size := (endTime - startTime) / bucketDuration
	for i := int64(0); i < size; i++ {
		timeBucket, exist := groups[i]
		if !exist {
			timeBucket = make(map[string]int)
		}
		points = append(points, model.RootCauseCountPoint{
			Timestamp:         startTime/1e6 + step.Milliseconds()*i,
			RootCauseCountMap: timeBucket,
		})
	}
	return points, nil
}

func (c *ClickhouseAPI) queryRangeCausePerMin(
	tableName string,
	entryUri string,
	startTime int64,
	endTime int64,
	step time.Duration,
) (model.RootCauseCountTimeSeries, error) {
	bucketDuration := (step / 4).Nanoseconds()

	timeBucket := fmt.Sprintf("((end_time - %d) DIV %d)", startTime-bucketDuration, bucketDuration)

	fieldBuilder := NewFieldBuilder().
		Alias("labels['mutated_service']", "mutated_service").
		Alias(timeBucket, "time_bucket").
		Alias("COUNT(1)", "total")
	whereClause := NewQueryBuilder().
		Between("end_time", uint64(startTime-bucketDuration), uint64(endTime)).
		Equals("labels['content_key']", entryUri).
		Equals("is_drop", false)
	byBuilder := NewByLimitBuilder().
		GroupBy("mutated_service", "time_bucket")

	sql := fmt.Sprintf("SELECT %s FROM %s %s %s", fieldBuilder.String(), tableName, whereClause.String(), byBuilder.String())
	var results []struct {
		Service    string `ch:"mutated_service"`
		TimeBucket int64  `ch:"time_bucket"`
		Total      uint64 `ch:"total"`
	}
	if err := c.conn.Select(context.Background(), &results, sql, whereClause.values...); err != nil {
		log.Printf("Not found root cause counts: %v", err)
		return []model.RootCauseCountPoint{}, err
	}
	groups := map[int64]map[string]int{}
	for _, result := range results {
		timeBuckets := getMinuteTimeBuckets(result.TimeBucket)
		for _, timeBucket := range timeBuckets {
			serviceGroup, exist := groups[timeBucket]
			if !exist {
				serviceGroup = map[string]int{}
			}
			value, exist := serviceGroup[result.Service]
			if !exist {
				serviceGroup[result.Service] = int(result.Total)
			} else {
				serviceGroup[result.Service] = value + int(result.Total)
			}
			groups[timeBucket] = serviceGroup
		}
	}

	var points = make([]model.RootCauseCountPoint, 0)
	size := (endTime - startTime) / step.Nanoseconds()
	for i := int64(0); i < size; i++ {
		timeBucket, exist := groups[i]
		if !exist {
			timeBucket = make(map[string]int)
		}
		points = append(points, model.RootCauseCountPoint{
			Timestamp:         startTime/1e6 + step.Milliseconds()*i,
			RootCauseCountMap: timeBucket,
		})
	}
	return points, nil
}

// Bucket to Points
// 0,1,2,3,4 -> point1
// 4,5,6,7,8 -> point2
// 8,9,10,11,12 -> point3
// ...
func getMinuteTimeBuckets(index int64) []int64 {
	// 0 ~ 3 -> 0
	// 4 -> [0, 1]
	// 5 ~ 7 -> 1
	// 8 -> [1, 2]
	// ...
	intVal := index / 4
	if index%4 == 0 {
		if intVal == 0 {
			return []int64{0}
		} else {
			return []int64{intVal - 1, intVal}
		}
	} else {
		return []int64{intVal}
	}
}
