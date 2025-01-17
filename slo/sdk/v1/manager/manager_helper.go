package manager

import (
	"errors"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

func (m *SLORecordManager) getEntryTargets(service string, startTimeMill int64, endTimeMill int64, _ *api.PageParam) (targets []model.SLOTarget, err error) {
	activeEntries, err := m.ListEntry(service, startTimeMill, endTimeMill)
	if err != nil {
		return nil, err
	}
	// if err != nil {
	// 	targets = m.GetTargetsByKeys(aliasKeys)
	// 	return targets, nil
	// }

	// keys := aliasKeys
	// for i := 0; i < len(activeEntries); i++ {
	// 	if isDuplicatesTarget(keys, &activeEntries[i]) {
	// 		continue
	// 	}
	// 	keys = append(keys, activeEntries[i])
	// }

	targets = m.GetTargetsByKeys(activeEntries)
	return targets, nil
}

// func isDuplicatesTarget(targets []model.SLOEntryKey, newTarget *model.SLOEntryKey) bool {
// 	for i := 0; i < len(targets); i++ {
// 		if targets[i].EntryURI == newTarget.EntryURI {
// 			return true
// 		}
// 	}
// 	return false
// }

func (m *SLORecordManager) GetResultForTarget(
	entryTarget *model.SLOTarget,
	skipInactiveEntry bool,
	skipHealthyEntry bool,
	startUnixMilli int64,
	endUnixMilli int64,
	stepDuration time.Duration,
	checkSkipOnly bool,
) (res *model.SLOResult, isSkip bool, err error) {
	entryKey := entryTarget.InfoRef.KeyRef
	groups, err := m.GetTimeSeriesGroupResult(entryKey, entryTarget.SLOConfigs, startUnixMilli, endUnixMilli, stepDuration)

	if err != nil {
		if skipInactiveEntry && errors.Is(err, model.ErrNotActiveUriError) {
			return nil, true, nil
		}
	}

	if skipHealthyEntry {
		if isHealthyAllTime(groups) {
			return nil, true, err
		}
	}

	if checkSkipOnly {
		return nil, false, err
	}

	m.EnrichSLOGroup(groups, entryKey.EntryURI, startUnixMilli*1e6, endUnixMilli*1e6, stepDuration)

	return &model.SLOResult{
		SLOServiceName: model.SLOServiceName{
			EntryUri: entryKey.EntryURI,
			Alias:    entryTarget.InfoRef.Alias,
		},
		SLOGroup: groups,
	}, false, err
}

func isHealthyAllTime(groups []model.SLOGroup) bool {
	for _, group := range groups {
		if group.Status == model.NotAchieved {
			return false
		}
	}
	return true
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
