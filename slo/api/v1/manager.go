package api

import (
	"strings"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

type Manager interface {
	GetSLOResult(entryURL string, startUnixMilli int64, endUnixMilli int64, pageParam *PageParam, stepDuration time.Duration, skipInactiveEntry bool, skipHealthyEntry bool) (result []*model.SLOResult, count int, err error)
	SearchSLOResult(entryURL string, startMS int64, endMS int64, pageParam *PageParam, duration time.Duration, skipInactiveEntry bool, skipHealthyEntry bool, options ...SortByOption) (result []*model.SLOResult, count int, err error)
	GetAndStoreSLOResult(startUnixMilli int64, endUnixMilli int64, step time.Duration)
}

type SortByOption = string

const (
	requestCountSort     string = "requestCount"
	notAchievedCountSort string = "notAchievedCount"

	SortByRequestCount     SortByOption = "requestCount_total"
	SortByNotAchievedCount SortByOption = "notAchievedCount"
)

func GetSortOptions(sortOptions string) []SortByOption {
	var options = make([]SortByOption, 0)
	optionsStr := strings.Split(sortOptions, ",")
	for _, option := range optionsStr {
		switch option {
		case requestCountSort:
			options = append(options, SortByRequestCount)
		case notAchievedCountSort:
			options = append(options, SortByNotAchievedCount)
		}
	}
	return options
}

type PageParam struct {
	PageNum  int
	PageSize int
}

func GetPageParam(pageSize int, currentPage int) *PageParam {
	if pageSize < 1 || currentPage < 1 {
		return nil
	}
	return &PageParam{
		PageNum:  currentPage,
		PageSize: pageSize,
	}
}
