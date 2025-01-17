package model

import (
	"fmt"
	"strings"
)

type Traces struct {
	TraceId string
	Traces  []*Trace

	SentTraceCount   int
	UnSentTraceCount int
	MetricCount      int
	HasSlow          bool
	HasError         bool
	RootTrace        *Trace
}

func NewTraces(traceId string) *Traces {
	return &Traces{
		TraceId:          traceId,
		Traces:           make([]*Trace, 0),
		SentTraceCount:   0,
		UnSentTraceCount: 0,
	}
}

func (traces *Traces) AddTrace(trace *Trace) bool {
	if trace == nil {
		return false
	}
	traces.Traces = append(traces.Traces, trace)
	if trace.Labels.TopSpan {
		traces.RootTrace = trace
	}
	if trace.Labels.IsProfiled {
		if trace.Labels.IsSlowReport() {
			traces.HasSlow = true
		}
		if trace.Labels.IsErrorReport() {
			traces.HasError = true
		}
	}
	if trace.IsSent {
		traces.SentTraceCount += 1
	} else {
		traces.UnSentTraceCount += 1
	}

	return true
}

func (traces *Traces) HasSingleTrace() bool {
	for _, trace := range traces.Traces {
		if trace.Labels.IsSingleTrace() {
			return true
		}
	}
	return false
}

func (traces *Traces) HasChangedSample() bool {
	sample := -1
	for _, trace := range traces.Traces {
		if sample == -1 {
			sample = trace.Labels.SampleValue
		} else if sample != trace.Labels.SampleValue {
			return true
		}
	}
	return false
}

func (traces *Traces) GetQueryTrace() *Trace {
	if traces.RootTrace != nil {
		return traces.RootTrace
	}

	var entryTrace *Trace
	for _, trace := range traces.Traces {
		if entryTrace == nil || entryTrace.Labels.Duration < trace.Labels.Duration {
			entryTrace = trace
		}
	}
	return entryTrace
}

func (traces *Traces) FindTrace(spanId string) *Trace {
	for _, trace := range traces.Traces {
		if trace.Labels.ApmSpanId == spanId {
			return trace
		}
	}
	return nil
}

func (traces *Traces) GetTraceCount() int {
	return traces.SentTraceCount + traces.UnSentTraceCount
}

func (traces *Traces) ToString() string {
	var text strings.Builder
	text.WriteString(fmt.Sprintf("TraceId: %s, Traces: [\n", traces.TraceId))
	for i, trace := range traces.Traces {
		if i > 0 {
			text.WriteString(",\n")
		}
		text.WriteString(fmt.Sprintf("    {%s}", structToString(trace)))
	}
	text.WriteString("\n]")
	return text.String()
}

func (traces *Traces) GetSpanIdTraceMap() map[string]*Trace {
	spanIdTraceMap := make(map[string]*Trace, 0)
	for _, sampledTrace := range traces.Traces {
		spanIdTraceMap[sampledTrace.Labels.ApmSpanId] = sampledTrace
	}
	return spanIdTraceMap
}
