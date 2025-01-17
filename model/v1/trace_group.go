package model

type TraceGroup struct {
	Name      string       `json:"name"`
	Timestamp uint64       `json:"timestamp"`
	Version   string       `json:"data_version"`
	Source    string       `json:"data_source"`
	Labels    *TraceLabels `json:"labels"`
	Metrics   []Metric     `json:"metrics"`
	IsSent    bool         `json:"-"`
}

type ToTalTraces struct {
	SlowTraces  uint64 `json:"slowTraces"`
	ErrorTraces uint64 `json:"errorTraces"`
}

type Metric struct {
	Name string            `json:"Name"`
	Data map[string]uint64 `json:"Data"`
}
