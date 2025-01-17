package model

type CameraNodeReport struct {
	Timestamp uint64               `json:"timestamp"`
	TraceId   string               `json:"trace_id"`
	Duration  uint64               `json:"duration"`
	Data      CameraNodeReportData `json:"data"`
}

type CameraNodeReportData struct {
	EntryService        string `json:"entry_service"`
	EntryInstance       string `json:"entry_instance,omitempty"`
	MutatedService      string `json:"mutated_service"`
	MutatedInstance     string `json:"mutated_instance,omitempty"`
	MutatedUrl          string `json:"mutated_url,omitempty"`
	MutatedSpan         string `json:"span_id"`
	MutatedPod          string `json:"mutated_pod,omitempty"`
	MutatedPodNS        string `json:"mutated_pod_ns,omitempty"`
	MutatedWorkloadName string `json:"mutated_workload_name,omitempty"`
	MutatedWorkloadType string `json:"mutated_workload_type,omitempty"`
	Cause               string `json:"cause"`
	ContentKey          string `json:"content_key"`

	RelationTree    *TraceTreeNode   `json:"relation_trees"`
	OTelClientCalls []*ApmClientCall `json:"otel_client_calls"`

	ThresholdType     ThresholdType  `json:"threshold_type"`
	ThresholdValue    float64        `json:"threshold_value"`
	ThresholdRange    ThresholdRange `json:"threshold_range"`
	ThresholdMultiple float64        `json:"threshold_multiple"`

	// Deprecated: Use RelationTree
	Relation string `json:"relation,omitempty"`
	// Deprecated: Use OTelClientCalls
	ClientCalls string `json:"client_calls,omitempty"`
}
