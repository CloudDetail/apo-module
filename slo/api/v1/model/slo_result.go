package model

type SLOResult struct {
	SLOServiceName SLOServiceName `json:"serviceName"`
	SLOGroup       []SLOGroup     `json:"sloGroup"`
}

type SLOGroup struct {
	StartTime    int64     `json:"startTime"`
	EndTime      int64     `json:"endTime"`
	RequestCount int       `json:"requestCount"`
	Status       SLOStatus `json:"status"`
	SLOs         []SLO     `json:"SLOs"`

	SlowRootCauseCount  map[string]int `json:"slowRootCauseCount"`
	ErrorRootCauseCount map[string]int `json:"errorRootCauseCount"`
}

type SLOServiceName struct {
	EntryUri     string `json:"entryUri"`
	EntryService string `json:"entryService"`
	Alias        string `json:"alias"`
}

type SLOHistory struct {
	SuccessRate map[string]float64        `json:"successRate"`
	Latency     map[string]HistoryLatency `json:"latency"`
}

type HistoryLatency struct {
	Range string  `json:"range"`
	Value float64 `json:"value"`
}

type RootCauseCountTimeSeries []RootCauseCountPoint

type RootCauseCountPoint struct {
	Timestamp int64
	// RootCauseCountMap
	// key: service
	// value: how many times the service is root cause
	RootCauseCountMap map[string]int
}
