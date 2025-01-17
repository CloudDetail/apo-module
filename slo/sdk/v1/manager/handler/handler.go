package handler

import (
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

type SLOGroupHandler interface {
	EnrichSLOGroup(groups []model.SLOGroup, entryURI string, startNano, endNano int64, stepDuration time.Duration)
}

type RecordStorage interface {
	StoreSLOResult(results []*model.SLOResult, startTSMillis int64, step time.Duration)
	SearchSLOResult(entryURL string,
		startMS int64,
		endMS int64,
		pageParam *api.PageParam,
		duration time.Duration,
		skipInactiveEntry bool,
		skipHealthyEntry bool,
		options ...api.SortByOption) (result []*model.SLOResult, count int, err error)
}
