package manager

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/checker"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/clickhouse"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/config"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/config/dynamic"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/elasticsearch"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/manager/handler"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/pql"
	"github.com/robfig/cron/v3"
)

var _ api.Manager = &SLORecordManager{}

var DefaultManager api.Manager

var enableDebugInfo bool = false
var enableSLORecord bool = true

var configExpireTime time.Duration = -1
var pqlType string = ""

type SLORecordManager struct {
	api.ConfigManager
	api.Checker
	handler.SLOGroupHandler
	handler.RecordStorage

	*cron.Cron
}

func SetSLODebug(debug bool) {
	enableDebugInfo = debug
	dynamic.SetSLODebug(debug)
}

func SetSLORecord(enable bool) {
	enableSLORecord = enable
}

func SetPQLType(typeStr string) {
	pqlType = typeStr
}

func SetupConfigExpireTime(duration time.Duration) {
	configExpireTime = duration
}

func NewSLORecordManager(cfg *SLOManagerConfig) (*SLORecordManager, error) {
	if cfg == nil {
		return nil, errors.New("slo manager config can not be nil")
	}

	if !cfg.Enable {
		log.Printf("slo manager is disabled by config")
		return nil, nil
	}

	pqlApi, err := pql.NewPQLApi(cfg.Checker.PrometheusAddr, cfg.Checker.PQLType)
	if err != nil {
		log.Printf("failed to create PQLApi, can not analyzer slo correct!!!, err: %s", err)
		return nil, err
	}

	ckr := checker.NewPrometheusChecker(pqlApi)
	dynamicSLO := dynamic.NewDynamicDefaultSLOTarget(pqlApi, configExpireTime)
	configCache := config.NewSLOConfigCache(&cfg.CenterServer, dynamicSLO)
	dynamicSLO.SetupConfigCache(configCache)
	dynamicSLO.DynamicRecalculateStore()

	sloCron := cron.New(cron.WithParser(cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)))
	sloCron.AddFunc("@daily", dynamicSLO.DynamicRecalculateStore)
	sloCron.AddFunc("0 1 * * * *", dynamicSLO.Reset)
	manager := &SLORecordManager{
		Checker:       ckr,
		ConfigManager: configCache,
		Cron:          sloCron,
	}

	if cfg.EnableStorage {
		isStorageReady := false
		switch cfg.Storage.StorageType {
		case "elasticsearch":
			esAPI, err := elasticsearch.NewElasticSearchAPI(&cfg.Storage.Elasticsearch)
			if err != nil {
				return nil, fmt.Errorf("can not init esAPI for sloManager, err: %v", err)
			}
			manager.RecordStorage = esAPI
			manager.SLOGroupHandler = esAPI
			isStorageReady = true
		case "clickhouse":
			clickhouseAPI, err := clickhouse.NewClickhouseAPI(&cfg.Storage.Clickhouse)
			if err != nil {
				return nil, fmt.Errorf("can not init clickhouseAPI for sloManager, err: %v", err)
			}
			manager.RecordStorage = clickhouseAPI
			manager.SLOGroupHandler = clickhouseAPI
			isStorageReady = true
		default:
			log.Printf("storage type[%s] is not supported, disable SLORecord storage or query", cfg.Storage.StorageType)
		}

		if isStorageReady && cfg.EnableGenerateSLORecord {
			manager.Cron.AddFunc("30 * * * * *", manager.StoreLastMinuteSLORecords)
			manager.Cron.AddFunc("0 1 * * * *", manager.StoreLastHourSLORecords)
		}
	}

	manager.Cron.Start()
	DefaultManager = manager
	checker.DefaultChecker = ckr
	return manager, nil
}

func (m *SLORecordManager) GetSLOResult(service string, startUnixMilli int64, endUnixMilli int64, pageParam *api.PageParam, stepDuration time.Duration, skipInactiveEntry bool, skipHealthyEntry bool) (result []*model.SLOResult, count int, err error) {
	entryTargets, err := m.getEntryTargets(service, startUnixMilli, endUnixMilli, pageParam)
	if err != nil && len(entryTargets) == 0 {
		return nil, 0, err
	}

	result = make([]*model.SLOResult, 0)
	if pageParam == nil {
		if enableDebugInfo {
			log.Printf("Active entry in last five minutes:")
		}

		for i := 0; i < len(entryTargets); i++ {
			res, isSkip, err := m.GetResultForTarget(&entryTargets[i], skipInactiveEntry, skipHealthyEntry, startUnixMilli, endUnixMilli, stepDuration, false)

			if enableDebugInfo {
				if len(res.SLOGroup) == 0 {
					log.Printf("\t [%s]: no request in last one minute", entryTargets[i].InfoRef.KeyRef.EntryURI)
				}
				for _, group := range res.SLOGroup {
					log.Printf("\t [%s]: SLO: %+v", entryTargets[i].InfoRef.KeyRef.EntryURI, group)
				}
			}

			if err != nil && !errors.Is(err, model.ErrNotActiveUriError) {
				log.Printf("got error durning get slo result for [%s]: %s", entryTargets[i].InfoRef.KeyRef.EntryURI, err)
			}
			if isSkip || res == nil {
				continue
			}
			result = append(result, res)
		}
		return result, len(result), nil
	}

	from, to := pageByGolang(len(entryTargets), pageParam)

	if !skipHealthyEntry && !skipInactiveEntry {
		for i := from; i < to; i++ {
			res, _, err := m.GetResultForTarget(&entryTargets[i], skipInactiveEntry, skipHealthyEntry, startUnixMilli, endUnixMilli, stepDuration, false)
			if err != nil && !errors.Is(err, model.ErrNotActiveUriError) {
				log.Printf("got error durning get slo result for [%s]: %s", entryTargets[i].InfoRef.KeyRef.EntryURI, err)
			}
			result = append(result, res)
		}
		return result, len(entryTargets), nil
	}

	var checkSkipOnly = false
	var resultCount = 0

	for i := 0; i < len(entryTargets); i++ {
		res, isSkip, err := m.GetResultForTarget(&entryTargets[i], skipInactiveEntry, skipHealthyEntry, startUnixMilli, endUnixMilli, stepDuration, checkSkipOnly)
		if err != nil && !errors.Is(err, model.ErrNotActiveUriError) {
			log.Printf("got error durning get slo result for [%s]:", entryTargets[i].InfoRef.KeyRef.EntryURI)
		}
		if isSkip {
			continue
		}
		resultCount++
		if resultCount < from-1 || resultCount >= to {
			checkSkipOnly = true
			continue
		}
		if resultCount == from-1 {
			checkSkipOnly = false
			continue
		}
		result = append(result, res)
	}

	return result, resultCount, nil
}

func (m *SLORecordManager) GetAndStoreSLOResult(startUnixMilli int64, endUnixMilli int64, step time.Duration) {
	results, _, err := m.GetSLOResult("", startUnixMilli, endUnixMilli, nil, step, false, false)
	if err != nil {
		log.Printf("failed to get any slo result, err: %s", err)
		return
	}
	if len(results) == 0 {
		if enableDebugInfo {
			log.Printf("get empty slo result")
		}
		return
	}
	m.RecordStorage.StoreSLOResult(results, startUnixMilli, step)
}

func (m *SLORecordManager) StoreLastMinuteSLORecords() {
	now := time.Now()
	roundedTS := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
	m.GetAndStoreSLOResult(roundedTS.Add(-time.Minute).UnixMilli(), roundedTS.UnixMilli(), time.Minute)
}

func (m *SLORecordManager) StoreLastHourSLORecords() {
	now := time.Now()
	roundedTS := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
	m.GetAndStoreSLOResult(roundedTS.Add(-time.Hour).UnixMilli(), roundedTS.UnixMilli(), time.Hour)
}

func GetSLOResult(entryURL string, startUnixMilli int64, endUnixMilli int64, pageParam *api.PageParam, stepDuration time.Duration, skipInactiveEntry bool, skipHealthyEntry bool) (result []*model.SLOResult, count int, err error) {
	return DefaultManager.GetSLOResult(entryURL, startUnixMilli, endUnixMilli, pageParam, stepDuration, skipInactiveEntry, skipInactiveEntry)
}

func GetSLOFromCache(
	entryURL string,
	startUnixMilli int64,
	endUnixMilli int64,
	pageParam *api.PageParam,
	stepDuration time.Duration,
	skipInactiveEntry bool,
	skipHealthyEntry bool,
	options ...api.SortByOption,
) (result []*model.SLOResult, count int, err error) {
	if pageParam != nil {
		if pageParam.PageNum < 1 || pageParam.PageSize < 1 {
			return nil, 0, fmt.Errorf("invalid page param %+v", *pageParam)
		}
	}

	perResult, count, err := DefaultManager.SearchSLOResult(
		entryURL,
		startUnixMilli,
		endUnixMilli,
		pageParam,
		stepDuration,
		skipInactiveEntry,
		skipHealthyEntry,
		options...,
	)

	for _, result := range perResult {
		if config.DefaultConfigCache == nil {
			break
		}
		info := config.DefaultConfigCache.GetAlias(model.SLOEntryKey{EntryURI: result.SLOServiceName.EntryUri})
		if info == nil {
			result.SLOServiceName.Alias = ""
		} else {
			result.SLOServiceName.Alias = info.Alias
		}
	}

	return perResult, count, err
}
