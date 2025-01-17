package config

import "github.com/CloudDetail/apo-module/slo/api/v1/model"

type StaticDefaultSLOConfig struct {
	DefaultConfigs []model.SLOConfig
}

func (s *StaticDefaultSLOConfig) GetDefaultSLOConfig(key model.SLOEntryKey) ([]model.SLOConfig, error) {
	return s.DefaultConfigs, nil
}
