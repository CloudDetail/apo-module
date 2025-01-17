package model

import "time"

type SLOResultRecord struct {
	SLOServiceName *SLOServiceName `json:"serviceName"`
	SLOGroup       *SLOGroup       `json:"sloGroup"`
	Step           RecordStep      `json:"step"`
	IndexTimestamp int64           `json:"indexTimestamp"`
}

type RecordStep string

const (
	MinuteStep RecordStep = "minute"
	HourStep   RecordStep = "hour"
)

func GetRecordStepFromDuration(duration time.Duration) RecordStep {
	if duration == time.Minute {
		return MinuteStep
	} else if duration == time.Hour {
		return HourStep
	} else {
		return "unknown"
	}
}
