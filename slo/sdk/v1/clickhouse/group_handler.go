package clickhouse

import (
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/manager/handler"
)

var _ handler.SLOGroupHandler = &ClickhouseAPI{}

const (
	TableSlowReport  string = "slow_report"
	TableErrorReport string = "error_report"
)

// EnrichSLOGroup implements handler.SLOGroupHandler.
func (cAPI *ClickhouseAPI) EnrichSLOGroup(groups []model.SLOGroup, entryURI string, startNano int64, endNano int64, stepDuration time.Duration) {
	cAPI.addRootCauseCount(TableSlowReport, entryURI, startNano, endNano, stepDuration, groups)
	cAPI.addRootCauseCount(TableErrorReport, entryURI, startNano, endNano, stepDuration, groups)
}

func (cAPI *ClickhouseAPI) addRootCauseCount(tableName string, entryURI string, startNano, endNano int64, stepDuration time.Duration, groups []model.SLOGroup) {
	timeSeries, err := cAPI.QueryTimeSeriesRootCauseCount(tableName, &entryURI, startNano, endNano, stepDuration)
	if err != nil {
		for i := range groups {
			setRootCauseCount(tableName, &groups[i], map[string]int{})
		}
		return
	}

	var groupIdx int = 0
	for i := range timeSeries {
		timeSeriesTS := timeSeries[i].Timestamp
		groupIdx = skipNotMatchedGroup(groupIdx, groups, timeSeriesTS, stepDuration, tableName)
		if groupIdx >= len(groups) {
			break
		}

		groupsTS := groups[groupIdx].StartTime
		if groupsTS >= timeSeriesTS && groupsTS-timeSeriesTS < 60*int64(time.Second/time.Millisecond) ||
			(groupsTS < timeSeriesTS && timeSeriesTS-groupsTS < 60*int64(time.Second/time.Millisecond)) {
			setRootCauseCount(tableName, &groups[groupIdx], timeSeries[i].RootCauseCountMap)
			groupIdx++
		}
	}
}

func skipNotMatchedGroup(groupIdx int, groups []model.SLOGroup, timeSeriesTS int64, stepDuration time.Duration, index string) int {
	for ; groupIdx < len(groups); groupIdx++ {
		groupsTS := groups[groupIdx].StartTime
		if groupsTS+stepDuration.Milliseconds() > timeSeriesTS && groups[groupIdx].Status != model.Achieved {
			return groupIdx
		}
		setRootCauseCount(index, &groups[groupIdx], map[string]int{})
	}

	return groupIdx
}

func setRootCauseCount(tableName string, group *model.SLOGroup, counts map[string]int) {
	switch tableName {
	case TableSlowReport:
		group.SlowRootCauseCount = counts
	case TableErrorReport:
		group.ErrorRootCauseCount = counts
	}
}
