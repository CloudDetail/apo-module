package dynamic

import (
	"log"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/pql"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const staticDefaultLatencyExpected = 500 * time.Millisecond
const staticDefaultSuccessRateExpected = 95
const defaultLatencyMultiple = 1.1

var enableDebugInfo bool = false

func SetSLODebug(debug bool) {
	enableDebugInfo = debug
}

type DynamicDefaultSLOTarget struct {
	lruMap           LRUMap
	todayStartTSNano int64
	expireTime       time.Duration

	PQLApi pql.PQLApi
	store  api.ConfigManager
}

func NewDynamicDefaultSLOTarget(pqlAPI pql.PQLApi, expireTime time.Duration) *DynamicDefaultSLOTarget {
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	todayStartTSNano := today.UnixNano()

	var m LRUMap
	if expireTime > 0 {
		m = expirable.NewLRU[string, []model.SLOConfig](1e4, nil, expireTime)
	} else {
		m, _ = lru.New[string, []model.SLOConfig](1e4)
	}

	return &DynamicDefaultSLOTarget{
		expireTime:       expireTime,
		lruMap:           m,
		todayStartTSNano: todayStartTSNano,
		PQLApi:           pqlAPI,
	}
}

func (ddst *DynamicDefaultSLOTarget) GetDefaultSLOConfig(key model.SLOEntryKey) ([]model.SLOConfig, error) {
	configs, ok := ddst.lruMap.Get(key.EntryURI)
	if ok {
		return configs, nil
	}
	originalExpectedValue, source := ddst.getYesterdayLatencyOrLastOneHourOrDefault(0.9, key.EntryURI, ddst.todayStartTSNano, time.Now().UnixNano(), float64(staticDefaultLatencyExpected))
	defaultConfigs := []model.SLOConfig{
		{Type: model.SLO_LATENCY_P90_TYPE, Multiple: defaultLatencyMultiple, ExpectedValue: originalExpectedValue * defaultLatencyMultiple, Source: source},
		{Type: model.SLO_SUCCESS_RATE_TYPE, Multiple: 1.0, ExpectedValue: staticDefaultSuccessRateExpected, Source: model.ConstantExpectSource},
	}
	ddst.lruMap.Add(key.EntryURI, defaultConfigs)

	if enableDebugInfo {
		log.Printf("[SLO Setup] %s: SLO Target calculated by DynamicSLO: %+v", key.EntryURI, defaultConfigs)
	}
	return defaultConfigs, nil
}

func (ddst *DynamicDefaultSLOTarget) GetDefaultSLOConfigLastHour(key model.SLOEntryKey) ([]model.SLOConfig, error) {
	configs, ok := ddst.lruMap.Get(key.EntryURI)
	if ok {
		return configs, nil
	}
	originalExpectedValue, source := ddst.getLastOneHourOrDefault(0.9, key.EntryURI, time.Now().UnixNano(), float64(staticDefaultLatencyExpected))
	sloConfigs := []model.SLOConfig{
		{Type: model.SLO_LATENCY_P90_TYPE, Multiple: defaultLatencyMultiple, ExpectedValue: originalExpectedValue * defaultLatencyMultiple, Source: source},
		{Type: model.SLO_SUCCESS_RATE_TYPE, Multiple: 1.0, ExpectedValue: staticDefaultSuccessRateExpected, Source: model.ConstantExpectSource},
	}

	if source != model.DefaultExpectSource {
		ddst.lruMap.Add(key.EntryURI, sloConfigs)
	}

	if enableDebugInfo {
		log.Printf("[SLO Setup] %s: SLO Target calculated by DynamicSLO: %+v", key.EntryURI, sloConfigs)
	}
	return sloConfigs, nil
}

func (ddst *DynamicDefaultSLOTarget) Reset() {
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	todayStartTSNano := today.UnixNano()

	var m LRUMap
	if ddst.expireTime > 0 {
		m = expirable.NewLRU[string, []model.SLOConfig](1e4, nil, ddst.expireTime)
	} else {
		m, _ = lru.New[string, []model.SLOConfig](1e4)
	}

	ddst.lruMap = m
	ddst.todayStartTSNano = todayStartTSNano
	log.Printf("[SLO Setup] Reset Default SLO Target")
}
