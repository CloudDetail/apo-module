package manager

import (
	"log"
	"net/http"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/checker"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/config"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/config/dynamic"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/elasticsearch"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/manager/handler"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/pql"
	es7 "github.com/olivere/elastic/v7"
	"github.com/robfig/cron/v3"
)

func InitDefaultSLORecordManager(
	centerServerAddr string,
	portalClient *http.Client,
	promAddr string,
	esClient *es7.Client,
	esIndexSuffix string,
) api.Manager {
	configCache := config.NewSLOConfigCacheOld(centerServerAddr, portalClient)
	config.DefaultConfigCache = configCache

	pqlApi, err := pql.NewPQLApi(promAddr, pqlType)
	if err != nil {
		log.Printf("failed to create PQLApi, can not analyzer slo correct!!!, err: %s", err)
	}
	pChecker := checker.NewPrometheusChecker(pqlApi)
	checker.DefaultChecker = pChecker

	dynamicSLO := dynamic.NewDynamicDefaultSLOTarget(pqlApi, configExpireTime)
	dynamicSLO.SetupConfigCache(configCache)
	configCache.DefaultSLOConfig = dynamicSLO
	dynamicSLO.DynamicRecalculateStore()

	esStorage := elasticsearch.NewElasticElasticsearchAPIsearchAPI(esClient, esIndexSuffix)

	rm := newSLORecordManagerBuilder().
		withConfigManager(configCache).
		withChecker(pChecker).
		withSLOGroupHandler(esStorage).
		WithRecordStorage(esStorage)

	if rm.Cron == nil {
		rm.Cron = cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)))
	}
	rm.Cron.AddFunc("@daily", dynamicSLO.DynamicRecalculateStore)
	rm.Cron.AddFunc("0 1 * * * *", dynamicSLO.Reset)

	if enableSLORecord {
		rm.Cron.AddFunc("30 * * * * *", rm.StoreLastMinuteSLORecords)
		rm.Cron.AddFunc("0 1 * * * *", rm.StoreLastHourSLORecords)
	}

	rm.Cron.Start()

	DefaultManager = rm

	return rm
}

func InitDefaultSLOConfigCache(
	centerServerAddr string,
	portalClient *http.Client,
	promAddr string,
) api.Manager {
	configCache := config.NewSLOConfigCacheOld(centerServerAddr, portalClient)
	config.DefaultConfigCache = configCache

	pqlApi, err := pql.NewPQLApi(promAddr, pqlType)
	if err != nil {
		log.Printf("failed to create PQLApi, can not analyzer slo correct!!!, err: %s", err)
	}

	dynamicSLO := dynamic.NewDynamicDefaultSLOTarget(pqlApi, configExpireTime)
	dynamicSLO.SetupConfigCache(configCache)
	configCache.DefaultSLOConfig = dynamicSLO

	pChecker := checker.NewPrometheusChecker(pqlApi)
	checker.DefaultChecker = pChecker

	rm := newSLORecordManagerBuilder().withConfigManager(configCache)

	if rm.Cron == nil {
		rm.Cron = cron.New(cron.WithParser(cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		)))
	}
	rm.Cron.AddFunc("@daily", dynamicSLO.DynamicRecalculateStore)
	rm.Cron.AddFunc("0 1 * * * *", dynamicSLO.Reset)
	rm.Cron.Start()

	DefaultManager = rm

	return rm
}

func (m *SLORecordManager) withConfigManager(configManager api.ConfigManager) *SLORecordManager {
	m.ConfigManager = configManager
	return m
}

func (m *SLORecordManager) withChecker(checker api.Checker) *SLORecordManager {
	m.Checker = checker
	return m
}

func (m *SLORecordManager) withSLOGroupHandler(handler handler.SLOGroupHandler) *SLORecordManager {
	m.SLOGroupHandler = handler
	return m
}

func (m *SLORecordManager) WithRecordStorage(storage handler.RecordStorage) *SLORecordManager {
	m.RecordStorage = storage
	return m
}

func newSLORecordManagerBuilder() *SLORecordManager {
	return &SLORecordManager{}
}
