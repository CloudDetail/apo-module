package model

import "fmt"

type Trace struct {
	Timestamp        uint64       `json:"timestamp"`
	Version          string       `json:"data_version"`
	Source           string       `json:"data_source"`
	Labels           *TraceLabels `json:"labels"`
	WorkloadName     string       `json:"workload_name"`
	WorkloadKind     string       `json:"workload_kind"`
	PodIp            string       `json:"pod_ip"`
	PodName          string       `json:"pod_name"`
	Namespace        string       `json:"namespace"`
	OnOffMetrics     string       `json:"onoff_metrics"`
	BaseOnOffMetrics string       `json:"base_onoff_metrics"`
	BaseRange        string       `json:"base_range"`
	MutatedType      string       `json:"mutated_type"`
	IsSent           bool         `json:"-"`
}

func (trace *Trace) GetInstanceId() string {
	if len(trace.PodName) > 0 {
		return trace.PodName
	}
	if len(trace.Labels.ContainerId) > 0 {
		return fmt.Sprintf("%s@%s@%s", trace.Labels.ServiceName, trace.Labels.NodeName, trace.Labels.ContainerId)
	}
	if trace.Labels.Pid > 0 {
		return fmt.Sprintf("%s@%s@%d", trace.Labels.ServiceName, trace.Labels.NodeName, trace.Labels.Pid)
	}
	return trace.Labels.ServiceName
}

func (trace *Trace) SetOnOffMetrics(onOffMetrics string) {
	trace.OnOffMetrics = onOffMetrics
}

func (trace *Trace) MarkSent() {
	trace.IsSent = true
}

type TraceLabels struct {
	Pid               uint32         `json:"pid"`
	Tid               uint32         `json:"tid"`
	TopSpan           bool           `json:"top_span"`
	Protocol          string         `json:"protocol"`
	ServiceName       string         `json:"service_name"`
	Url               string         `json:"content_key"`
	HttpUrl           string         `json:"http_url"`
	IsSilent          bool           `json:"is_silent"`
	IsSampled         bool           `json:"is_sampled"`
	IsSlow            bool           `json:"is_slow"`
	IsServer          bool           `json:"is_server"`
	IsError           bool           `json:"is_error"`
	IsProfiled        bool           `json:"is_profiled"`
	SampleValue       int            `json:"sample_value"`
	ReportType        uint32         `json:"report_type"`
	ThresholdType     ThresholdType  `json:"threshold_type"`
	ThresholdValue    float64        `json:"threshold_value"`
	ThresholdRange    ThresholdRange `json:"threshold_range"`
	ThresholdMultiple float64        `json:"threshold_multiple"`
	TraceId           string         `json:"trace_id"`
	ApmType           string         `json:"apm_type"`
	ApmSpanId         string         `json:"apm_span_id"`
	Attributes        string         `json:"attributes"`
	ContainerId       string         `json:"container_id"`
	ContainerName     string         `json:"container_name"`
	StartTime         uint64         `json:"start_time"`
	Duration          uint64         `json:"duration"`
	EndTime           uint64         `json:"end_time"`
	NodeName          string         `json:"node_name"`
	NodeIp            string         `json:"node_ip"`
	OffsetTs          int64          `json:"offset_ts"`
}

func (trace *TraceLabels) IsSlowReport() bool {
	return trace.ReportType == 1 || trace.ReportType == 4
}

func (trace *TraceLabels) IsNormalReport() bool {
	return trace.ReportType == 2
}

func (trace *TraceLabels) IsErrorReport() bool {
	return trace.ReportType == 3 || trace.ReportType == 4
}

func (trace *TraceLabels) IsSlowAndErrorReport() bool {
	return trace.ReportType == 4
}

func (trace *TraceLabels) IsSingleTrace() bool {
	return trace.ReportType > 4
}
