package dynamic

import (
	"log"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/pql"
)

const (
	dayDuration  = "24h"
	hourDuration = "1h"
)

func (ddst *DynamicDefaultSLOTarget) SetupConfigCache(storeSource api.ConfigManager) {
	ddst.store = storeSource
}

func (ddst *DynamicDefaultSLOTarget) DynamicRecalculateStore() {
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	todayStartTSNano := today.UnixNano()
	nowTSNano := now.UnixNano()

	targetMap := ddst.store.ListTarget()
	targetMap.Range(func(keyRef, targetInfoRef any) bool {
		key := keyRef.(model.SLOEntryKey)
		targetInfo := targetInfoRef.(*model.SLOTarget)
		configs := targetInfo.SLOConfigs
		for i := 0; i < len(configs); i++ {
			if configs[i].Source == model.ConstantExpectSource {
				continue
			}

			if !model.IsLatencyPercentileSLOType(configs[i].Type) {
				configs[i].Source = model.ConstantExpectSource
				continue
			}
			percentile := model.GetLatencyPercentileByType(configs[i].Type)
			latencyValue, source := ddst.getYesterdayLatencyOrLastOneHourOrDefault(percentile, key.EntryURI, todayStartTSNano, nowTSNano, float64(staticDefaultLatencyExpected))
			configs[i].Source = source
			configs[i].ExpectedValue = latencyValue * configs[i].Multiple
		}

		if enableDebugInfo {
			log.Printf("[SLO Setup] %s: SLO Target calculated by DynamicSLO: %+v", key.EntryURI, configs)
		}
		return true
	})
}

func (ddst *DynamicDefaultSLOTarget) getYesterdayLatencyOrLastOneHourOrDefault(percentile float64, entryURI string, todayTSNano int64, nowTSNano int64, defaultValue float64) (float64, model.ExpectedSource) {
	if value, err := ddst.PQLApi.QueryMetric(uint64(todayTSNano), pql.GetLatencyPercentilePQL(percentile, entryURI, dayDuration, ddst.PQLApi.BucketLabelName())); err == nil && value != 0 {
		return value / 1e6, model.YesterdayExpectSource
	}
	if value, err := ddst.PQLApi.QueryMetric(uint64(nowTSNano), pql.GetLatencyPercentilePQL(percentile, entryURI, hourDuration, ddst.PQLApi.BucketLabelName())); err == nil && value != 0 {
		return value / 1e6, model.LastHourExpectedSource
	}
	return defaultValue / 1e6, model.DefaultExpectSource
}

func (ddst *DynamicDefaultSLOTarget) getLastOneHourOrDefault(percentile float64, entryURI string, nowTSNano int64, defaultValue float64) (float64, model.ExpectedSource) {
	if value, err := ddst.PQLApi.QueryMetric(uint64(nowTSNano), pql.GetLatencyPercentilePQL(percentile, entryURI, hourDuration, ddst.PQLApi.BucketLabelName())); err == nil && value != 0 {
		return value / 1e6, model.LastHourExpectedSource
	}
	return defaultValue / 1e6, model.DefaultExpectSource
}
