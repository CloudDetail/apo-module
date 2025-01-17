package api

import (
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

type Checker interface {
	ListEntryTemp(service string, startTimeMill int64, endTimeMill int64) (targets []model.SLOEntryKeyTemp, err error)

	ListEntry(service string, startTimeMill int64, endTimeMill int64) (targets []model.SLOEntryKey, err error)

	ListContentKeyTemp(service string, startTimeMill int64, endTimeMill int64) (targets []model.SLOEntryKeyTemp, err error)

	GetTimeSeriesGroupResult(key *model.SLOEntryKey, sloConfigs []model.SLOConfig, startTime int64, endTime int64, step time.Duration) ([]model.SLOGroup, error)

	GetHistorySLO(key model.SLOEntryKey, endTime int64) (model.SLOHistory, error)
}
