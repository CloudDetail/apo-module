package model

type ErrorReport struct {
	Name      string           `json:"name"`
	Timestamp uint64           `json:"timestamp"`
	TraceId   string           `json:"trace_id"`
	IsDrop    bool             `json:"is_drop"`
	Duration  uint64           `json:"duration"`
	Data      *ErrorReportData `json:"data"`
}

type ErrorReportData struct {
	EntryService        string         `json:"entry_service,omitempty"`
	EntryInstance       string         `json:"entry_instance,omitempty"`
	MutatedService      string         `json:"mutated_service,omitempty"`
	MutatedInstance     string         `json:"mutated_instance,omitempty"`
	MutatedUrl          string         `json:"mutated_url,omitempty"`
	MutatedSpan         string         `json:"span_id,omitempty"`
	MutatedPod          string         `json:"mutated_pod,omitempty"`
	MutatedPodNS        string         `json:"mutated_pod_ns,omitempty"`
	MutatedWorkloadName string         `json:"mutated_workload_name,omitempty"`
	MutatedWorkloadType string         `json:"mutated_workload_type,omitempty"`
	ContentKey          string         `json:"content_key,omitempty"`
	Cause               string         `json:"cause,omitempty"`
	CauseMessage        string         `json:"cause_message,omitempty"`
	RelationTree        *ErrorTreeNode `json:"relation_trees"`

	ThresholdType     ThresholdType  `json:"threshold_type"`
	ThresholdValue    float64        `json:"threshold_value"`
	ThresholdRange    ThresholdRange `json:"threshold_range"`
	ThresholdMultiple float64        `json:"threshold_multiple"`
}
