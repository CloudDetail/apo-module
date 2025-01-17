package dynamic

import "github.com/CloudDetail/apo-module/slo/api/v1/model"

type LRUMap interface {
	Add(key string, value []model.SLOConfig) (evicted bool)
	Get(key string) ([]model.SLOConfig, bool)
}
