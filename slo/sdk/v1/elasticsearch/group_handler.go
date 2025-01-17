package elasticsearch

import (
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/manager/handler"
)

var _ handler.SLOGroupHandler = &ElasticsearchAPI{}

const (
	CameraNodeReportIndex  string = "camera_node_report"
	CameraErrorReportIndex string = "camera_error_report"
)

func (esapi *ElasticsearchAPI) EnrichSLOGroup(groups []model.SLOGroup, entryURI string, startNano, endNano int64, stepDuration time.Duration) {
	esapi.addRootCauseCount(CameraNodeReportIndex, entryURI, startNano, endNano, stepDuration, groups)
	esapi.addRootCauseCount(CameraErrorReportIndex, entryURI, startNano, endNano, stepDuration, groups)
}

func (esapi *ElasticsearchAPI) addRootCauseCount(index string, entryURI string, startNano, endNano int64, stepDuration time.Duration, groups []model.SLOGroup) {
	timeSeries, err := esapi.QueryTimeSeriesRootCauseCount(index, &entryURI, startNano, endNano, stepDuration)
	if err != nil {
		for i := range groups {
			setRootCauseCount(index, &groups[i], map[string]int{})
		}
		return
	}

	var groupIdx int = 0
	for i := range timeSeries {
		timeSeriesTS := timeSeries[i].Timestamp
		groupIdx = esapi.skipNotMatchedGroup(groupIdx, groups, timeSeriesTS, stepDuration, index)
		if groupIdx >= len(groups) {
			break
		}

		groupsTS := groups[groupIdx].StartTime
		if groupsTS >= timeSeriesTS && groupsTS-timeSeriesTS < 60*int64(time.Second/time.Millisecond) ||
			(groupsTS < timeSeriesTS && timeSeriesTS-groupsTS < 60*int64(time.Second/time.Millisecond)) {
			setRootCauseCount(index, &groups[groupIdx], timeSeries[i].RootCauseCountMap)
			groupIdx++
		}
	}
}

func (esapi *ElasticsearchAPI) skipNotMatchedGroup(groupIdx int, groups []model.SLOGroup, timeSeriesTS int64, stepDuration time.Duration, index string) int {
	for ; groupIdx < len(groups); groupIdx++ {
		groupsTS := groups[groupIdx].StartTime
		if groupsTS+stepDuration.Milliseconds() > timeSeriesTS && groups[groupIdx].Status != model.Achieved {
			return groupIdx
		}
		setRootCauseCount(index, &groups[groupIdx], map[string]int{})
	}

	return groupIdx
}

func setRootCauseCount(index string, group *model.SLOGroup, counts map[string]int) {
	switch index {
	case CameraNodeReportIndex:
		group.SlowRootCauseCount = counts
	case CameraErrorReportIndex:
		group.ErrorRootCauseCount = counts
	}
}
