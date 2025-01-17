package checker

import (
	"math"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

type SLOTimeSeries struct {
	Res            []model.SLOGroup
	StartTimestamp int64
	EndTimestamp   int64
	StepMills      int64

	lastUpdateIdx int64
}

func NewSLOTimeSeries(startTimestamp int64, endTimestamp int64, step time.Duration) *SLOTimeSeries {
	var sloGroups = make([]model.SLOGroup, 0)
	var stepMills = step.Milliseconds()

	startTimestamp = startTimestamp - startTimestamp%stepMills
	endTimestamp = endTimestamp - endTimestamp%stepMills

	for tmpTS := startTimestamp; tmpTS < endTimestamp; tmpTS += stepMills {
		sloGroup := model.SLOGroup{
			StartTime:    tmpTS,
			EndTime:      tmpTS + stepMills,
			Status:       model.Unknown,
			SLOs:         make([]model.SLO, 0),
			RequestCount: 0,
		}
		sloGroups = append(sloGroups, sloGroup)
	}

	return &SLOTimeSeries{
		Res:            sloGroups,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		StepMills:      stepMills,
		lastUpdateIdx:  -1,
	}
}

func (slots *SLOTimeSeries) FinishUpdateSLO() {
	size := int64(len(slots.Res))
	if slots.lastUpdateIdx+1 < size {
		for i := slots.lastUpdateIdx + 1; i < size; i++ {
			slots.Res[i].Status = model.Unknown
		}
	}
	slots.lastUpdateIdx = -1
}

func (slots *SLOTimeSeries) AddSLOAtTimestamp(slo *model.SLO, ts int64) {
	if ts > slots.EndTimestamp || ts < slots.StartTimestamp {
		// JUST Drop it
		return
	}

	index := (ts-slots.StartTimestamp)/slots.StepMills - 1
	if slots.lastUpdateIdx+1 < index {
		for i := slots.lastUpdateIdx + 1; i < index; i++ {
			slots.Res[i].Status = model.Unknown
		}
	}

	var oldStatus = slots.Res[index].Status
	if len(slots.Res[index].SLOs) == 0 {
		oldStatus = model.Achieved
	}

	slots.Res[index].SLOs = append(slots.Res[index].SLOs, *slo)
	slots.Res[index].Status = MergeStatus(oldStatus, slo.Status)

	slots.lastUpdateIdx = index
}

func (slots *SLOTimeSeries) AddRequestCountAtTimestamp(requestCount float64, ts int64) {
	if ts > slots.EndTimestamp || ts < slots.StartTimestamp {
		// JUST Drop it
		return
	}
	index := (ts-slots.StartTimestamp)/slots.StepMills - 1
	slots.Res[index].RequestCount = int(requestCount)
}

func (slots *SLOTimeSeries) AllRequestCountFailed(sloConfig *model.SLOConfig) {
	for i := 0; i < len(slots.Res); i++ {
		if slots.Res[i].RequestCount > 0 {
			slots.Res[i].SLOs = append(slots.Res[i].SLOs, model.SLO{
				SLOConfig:    sloConfig,
				CurrentValue: 0,
				Status:       model.NotAchieved,
			})
			var oldStatus = slots.Res[i].Status
			if len(slots.Res[i].SLOs) == 0 {
				oldStatus = model.Achieved
			}
			slots.Res[i].Status = MergeStatus(oldStatus, model.NotAchieved)
		}
	}
}

func checkResult(c model.SLOConfig, value float64) *model.SLO {
	if math.IsNaN(value) {
		return &model.SLO{
			SLOConfig:    &c,
			CurrentValue: -1,
			Status:       model.Unknown,
		}
	}

	status := model.NotAchieved
	var currentValue float64
	if c.Type == model.SLO_SUCCESS_RATE_TYPE {
		currentValue = value
		if currentValue >= c.ExpectedValue {
			status = model.Achieved
		}
	} else if c.Type == model.SLO_LATENCY_P90_TYPE || c.Type == model.SLO_LATENCY_P95_TYPE || c.Type == model.SLO_LATENCY_P99_TYPE {
		currentValue = value / 1e6
		if currentValue <= c.ExpectedValue {
			status = model.Achieved
		}
	}

	return &model.SLO{
		SLOConfig:    &c,
		CurrentValue: currentValue,
		Status:       status,
	}
}

func getDayUnixMilli(unixMilli int64) int64 {
	i := time.UnixMilli(unixMilli)
	todayStart := time.Date(i.Year(), i.Month(), i.Day(), 0, 0, 0, 0, i.Location())
	return todayStart.UnixMilli()
}

func MergeStatus(oldStatus model.SLOStatus, newStatus model.SLOStatus) model.SLOStatus {
	switch {
	case newStatus == model.Unknown || newStatus == model.Achieved:
		return oldStatus
	case newStatus == model.NotAchieved:
		return model.NotAchieved
	default:
		return model.Unknown
	}
}
