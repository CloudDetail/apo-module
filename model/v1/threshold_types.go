package model

import "time"

type ThresholdType string

const (
	P90ThresholdType ThresholdType = "LatencyP90"
	P95ThresholdType ThresholdType = "LatencyP95"
	P99ThresholdType ThresholdType = "LatencyP99"
)

type ThresholdRange string

const (
	RangeLast1h    ThresholdRange = "last1h"
	RangeYesterday ThresholdRange = "yesterday"
	RangeConstant  ThresholdRange = "constant"
)

func (t ThresholdType) GetPercentile() float64 {
	switch t {
	case P90ThresholdType:
		return 0.9
	case P95ThresholdType:
		return 0.95
	case P99ThresholdType:
		return 0.99
	default:
		return 0
	}
}

func (t ThresholdType) GetPercentileString() string {
	switch t {
	case P90ThresholdType:
		return "0.9"
	case P95ThresholdType:
		return "0.95"
	case P99ThresholdType:
		return "0.99"
	default:
		return "0"
	}
}

func (r ThresholdRange) GetRange(traceStartTSNano int64) (startTSMill int64, duration string) {
	switch r {
	case RangeLast1h:
		return traceStartTSNano / 1e6, "1h"
	case RangeYesterday:
		startTime := time.UnixMilli(traceStartTSNano / 1e6)
		return time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, startTime.Location()).UnixMilli(), "24h"
	case RangeConstant:
		return traceStartTSNano / 1e6, "1h"
	default:
		return traceStartTSNano / 1e6, "1h"
	}
}

func (r ThresholdRange) GetDuration() (duration string) {
	switch r {
	case RangeLast1h:
		return "1h"
	case RangeYesterday:
		return "24h"
	case RangeConstant:
		return "1h"
	default:
		return "1h"
	}
}
