package api

import (
	"sync"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

type ConfigManager interface {
	ListAlias() map[string]*model.SLOEntryInfo
	GetAlias(key model.SLOEntryKey) *model.SLOEntryInfo
	GetCustomTarget(service string) []model.SLOEntryKey
	AddOrUpdateAlias(key model.SLOEntryKey, alias string)
	ListSLOConfig() map[string][]model.SLOConfig
	GetSLOConfig(key model.SLOEntryKey) []model.SLOConfig
	GetSLOConfigOrDefault(key model.SLOEntryKey) []model.SLOConfig
	GetSLOConfigOrDefaultInLastHour(key model.SLOEntryKey) []model.SLOConfig
	SetDefaultConfig(defaultConfigs DefaultSLOConfig)
	ListTarget() *sync.Map
	GetTargetsByKeys(keys []model.SLOEntryKey) []model.SLOTarget
	AddOrUpdateSLOTarget(key model.SLOEntryKey, configs []model.SLOConfig)
	DefaultSLOConfig
}

type DefaultSLOConfig interface {
	GetDefaultSLOConfig(model.SLOEntryKey) ([]model.SLOConfig, error)
	GetDefaultSLOConfigLastHour(model.SLOEntryKey) ([]model.SLOConfig, error)
}
