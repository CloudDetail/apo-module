package config

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

var _ api.ConfigManager = &ConfigCache{}

var DefaultConfigCache api.ConfigManager

type ConfigCache struct {
	SLOAliasMap sync.Map

	SLOTargetMap sync.Map

	api.DefaultSLOConfig

	centerServerAddr string
	client           *http.Client
}

// Setup implements SLOInfo.
func NewSLOConfigCacheOld(centerServerAddr string, client *http.Client) *ConfigCache {
	cache := &ConfigCache{
		centerServerAddr: centerServerAddr,
		client:           client,
	}

	cache.setupAliasMap()
	cache.setupTargetMap()

	return cache
}

func NewSLOConfigCache(cfg *CenterServerConfig, defaultSLOConfig api.DefaultSLOConfig) *ConfigCache {
	if cfg == nil {
		// impossible branch
		cache := &ConfigCache{}
		DefaultConfigCache = cache
		return cache
	}

	proxyClient := createHttpClient(cfg.ProxyAddress)
	cache := &ConfigCache{
		centerServerAddr: cfg.Address,
		client:           proxyClient,
		DefaultSLOConfig: defaultSLOConfig,
	}

	cache.setupAliasMap()
	cache.setupTargetMap()

	DefaultConfigCache = cache
	return cache
}

type Response struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (scc *ConfigCache) setupAliasMap() {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s/config/uri-alias", scc.centerServerAddr), nil)
	if err != nil {
		log.Printf("[set alias map]error happened when requesting config/uri-alias: %v", err)
		return
	}
	resp, err := scc.client.Do(request)
	if err != nil {
		log.Printf("[SLO setup] failed to get SLOAlias from server[%s] error: %s", scc.centerServerAddr, err)
		return
	}
	defer resp.Body.Close()

	var response Response
	var aliasMap = make(map[string]*model.SLOEntryInfo)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[SLO setup] failed to read the response body: %v", err)
		return
	}
	if err = json.Unmarshal(body, &response); err != nil {
		log.Printf("[SLO setup] failed to decode SLOAlias Response from server[%s] error: %s Raw message: %s",
			scc.centerServerAddr, err, body)
		return
	}

	if response.Status != "success" {
		log.Printf("[SLO setup] failed to get SLOAlias from server[%s] error: %s", scc.centerServerAddr, response.Message)
		return
	}

	if err = json.Unmarshal(response.Data, &aliasMap); err != nil {
		log.Printf("[SLO setup] failed to decode SLOAlias from response,error: %s", err)
		return
	}

	for k, v := range aliasMap {
		if len(v.Alias) == 0 {
			continue
		}
		scc.AddOrUpdateAlias(model.SLOEntryKey{EntryURI: k}, v.Alias)
	}
}

func (scc *ConfigCache) setupTargetMap() {
	request, err := http.NewRequest("GET", fmt.Sprintf("http://%s/config/slo", scc.centerServerAddr), nil)
	if err != nil {
		log.Printf("[SLO setup]error happened when requesting config/uri-alias: %v", err)
		return
	}
	resp, err := scc.client.Do(request)
	if err != nil {
		log.Printf("[SLO setup] get SLOConfig from server[%s] error: %s", scc.centerServerAddr, err)
		return
	}
	defer resp.Body.Close()

	var response Response
	var configMap = make(map[string][]model.SLOConfig)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[SLO setup] failed to read the response body: %v", err)
		return
	}
	if err = json.Unmarshal(body, &response); err != nil {
		log.Printf("[SLO setup] failed to decode SLOConfigs Response from server[%s] error: %s Raw message: %s",
			scc.centerServerAddr, err, body)
		return
	}

	if response.Status != "success" {
		log.Printf("[SLO setup] failed to get SLOConfigs from server[%s] error: %s", scc.centerServerAddr, response.Message)
		return
	}

	if err = json.Unmarshal(response.Data, &configMap); err != nil {
		log.Printf("[SLO setup] failed to decode SLOConfigs from response,error: %s", err)
		return
	}

	for k, v := range configMap {
		scc.AddOrUpdateSLOTarget(model.SLOEntryKey{EntryURI: k}, v)
	}
}

func (scc *ConfigCache) GetCustomTarget(service string) []model.SLOEntryKey {
	var result = make([]model.SLOEntryKey, 0)
	scc.SLOAliasMap.Range(func(key, value interface{}) bool {
		if strings.HasPrefix(value.(*model.SLOEntryInfo).Alias, service) {
			result = append(result, key.(model.SLOEntryKey))
		}
		return true
	})
	return result
}

func (scc *ConfigCache) GetTargetsByKeys(keys []model.SLOEntryKey) []model.SLOTarget {
	var result = make([]model.SLOTarget, 0)

	for i := 0; i < len(keys); i++ {
		if target, ok := scc.SLOTargetMap.Load(keys[i]); ok {
			result = append(result, *target.(*model.SLOTarget))
		} else {
			defaultConfig, _ := scc.GetDefaultSLOConfig(keys[i])
			if info, ok := scc.SLOAliasMap.Load(keys[i]); ok {
				result = append(result, model.SLOTarget{
					InfoRef:    info.(*model.SLOEntryInfo),
					SLOConfigs: defaultConfig,
				})
			} else {
				result = append(result, model.SLOTarget{
					InfoRef: &model.SLOEntryInfo{
						KeyRef: &keys[i],
					},
					SLOConfigs: defaultConfig,
				})
			}
		}
	}

	return result
}

func (scc *ConfigCache) AddOrUpdateAlias(key model.SLOEntryKey, alias string) {
	if info, ok := scc.SLOAliasMap.Load(key); ok {
		info := info.(*model.SLOEntryInfo)
		info.Alias = alias
		return
	}

	if target, ok := scc.SLOTargetMap.Load(key); ok {
		target := target.(*model.SLOTarget)
		target.InfoRef.Alias = alias
		scc.SLOAliasMap.Store(key, target.InfoRef)
		return
	}

	info := &model.SLOEntryInfo{
		KeyRef: &key,
		Alias:  alias,
	}
	scc.SLOAliasMap.Store(key, info)
}

func (scc *ConfigCache) GetAlias(key model.SLOEntryKey) *model.SLOEntryInfo {
	if info, ok := scc.SLOAliasMap.Load(key); ok {
		return info.(*model.SLOEntryInfo)
	}
	return nil
}

// ListAlias implements SLOInfo.
func (scc *ConfigCache) ListAlias() map[string]*model.SLOEntryInfo {
	var res = make(map[string]*model.SLOEntryInfo)
	scc.SLOAliasMap.Range(func(key, value interface{}) bool {
		res[key.(model.SLOEntryKey).EntryURI] = value.(*model.SLOEntryInfo)
		return true
	})
	return res
}

func (scc *ConfigCache) AddOrUpdateSLOTarget(key model.SLOEntryKey, configs []model.SLOConfig) {
	if target, ok := scc.SLOTargetMap.Load(key); ok {
		target := target.(*model.SLOTarget)
		target.SLOConfigs = configs
		return
	}

	if info, ok := scc.SLOAliasMap.Load(key); ok {
		info := info.(*model.SLOEntryInfo)
		scc.SLOTargetMap.Store(key, &model.SLOTarget{
			InfoRef:    info,
			SLOConfigs: configs,
		})
		return
	}

	info := &model.SLOEntryInfo{
		KeyRef: &key,
	}
	scc.SLOAliasMap.Store(key, info)
	scc.SLOTargetMap.Store(key, &model.SLOTarget{
		InfoRef:    info,
		SLOConfigs: configs,
	})
}

func (scc *ConfigCache) GetSLOConfigOrDefault(key model.SLOEntryKey) []model.SLOConfig {
	if target, ok := scc.SLOTargetMap.Load(key); ok {
		return target.(*model.SLOTarget).SLOConfigs
	} else {
		defaultSLOConfig, _ := scc.GetDefaultSLOConfig(key)
		return defaultSLOConfig
	}
}

func (scc *ConfigCache) GetSLOConfigOrDefaultInLastHour(key model.SLOEntryKey) []model.SLOConfig {
	if target, ok := scc.SLOTargetMap.Load(key); ok {
		return target.(*model.SLOTarget).SLOConfigs
	} else {
		defaultSLOConfig, _ := scc.GetDefaultSLOConfigLastHour(key)
		return defaultSLOConfig
	}
}

// GetSLOConfig implements SLOInfo.
func (scc *ConfigCache) GetSLOConfig(key model.SLOEntryKey) []model.SLOConfig {
	if target, ok := scc.SLOTargetMap.Load(key); ok {
		return target.(*model.SLOTarget).SLOConfigs
	} else {
		return nil
	}
}

// ListSLOConfig implements SLOInfo.
func (scc *ConfigCache) ListSLOConfig() map[string][]model.SLOConfig {
	var res = make(map[string][]model.SLOConfig)
	scc.SLOTargetMap.Range(func(key, value interface{}) bool {
		target := value.(*model.SLOTarget)
		res[target.InfoRef.KeyRef.EntryURI] = target.SLOConfigs
		return true
	})
	return res
}

func (scc *ConfigCache) SetDefaultConfig(defaultConfigs api.DefaultSLOConfig) {
	scc.DefaultSLOConfig = defaultConfigs
}

func (scc *ConfigCache) ListTarget() *sync.Map {
	return &scc.SLOTargetMap
}

func GetTargetsByKeys(keys []model.SLOEntryKey) []model.SLOTarget {
	return DefaultConfigCache.GetTargetsByKeys(keys)
}

func GetCustomTarget(service string) []model.SLOEntryKey {
	return DefaultConfigCache.GetCustomTarget(service)
}

func AddOrUpdateAlias(key model.SLOEntryKey, alias string) {
	DefaultConfigCache.AddOrUpdateAlias(key, alias)
}

func GetAlias(key model.SLOEntryKey) *model.SLOEntryInfo {
	return DefaultConfigCache.GetAlias(key)
}

func ListAlias() map[string]*model.SLOEntryInfo {
	return DefaultConfigCache.ListAlias()
}

func GetSLOConfig(key model.SLOEntryKey) []model.SLOConfig {
	return DefaultConfigCache.GetSLOConfig(key)
}

func ListSLOConfig() map[string][]model.SLOConfig {
	return DefaultConfigCache.ListSLOConfig()
}

func GetSLOConfigOrDefault(key model.SLOEntryKey) []model.SLOConfig {
	return DefaultConfigCache.GetSLOConfigOrDefault(key)
}

func AddOrUpdateSLOTarget(key model.SLOEntryKey, configs []model.SLOConfig) {
	DefaultConfigCache.AddOrUpdateSLOTarget(key, configs)
}

func SetDefaultConfig(defaultConfigs api.DefaultSLOConfig) {
	DefaultConfigCache.SetDefaultConfig(defaultConfigs)
}
