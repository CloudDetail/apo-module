package client

import (
	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/CloudDetail/apo-module/model/v1"
)

type NodeSpanTraces struct {
	Traces []*NodeSpanTrace
}

func NewNodeSpanTraces(apmType string, serviceNodes []*apmmodel.OtelServiceNode, sampledTraces *model.Traces) *NodeSpanTraces {
	spanTraces := &NodeSpanTraces{
		Traces: make([]*NodeSpanTrace, 0),
	}
	spanTraces.buildEntryTraces(apmType, serviceNodes, sampledTraces)
	return spanTraces
}

func (traces *NodeSpanTraces) buildEntryTraces(apmType string, serviceNodes []*apmmodel.OtelServiceNode, sampledTraces *model.Traces) {
	spanTraces := make(map[string]*model.Trace, 0)
	for _, sampledTrace := range sampledTraces.Traces {
		spanTraces[sampledTrace.Labels.ApmSpanId] = sampledTrace
	}

	for _, serviceNode := range serviceNodes {
		traces.collectApmTraceTree(apmType, serviceNode, spanTraces, nil)
	}
}

func (traces *NodeSpanTraces) collectApmTraceTree(apmType string, service *apmmodel.OtelServiceNode, sampledTraces map[string]*model.Trace, parent *NodeSpanTrace) {
	sampledSpanTrace := GetMatchSampledSpanTrace(apmType, service, sampledTraces)
	if sampledSpanTrace != nil {
		service.SetSpanId(sampledSpanTrace.Labels.ApmSpanId)
		if len(sampledSpanTrace.Labels.Attributes) > 0 {
			service.Attribute = sampledSpanTrace.Labels.Attributes
		}
	}

	for _, entrySpan := range service.EntrySpans {
		if entrySpan.PSpanId == "" {
			service.IsRoot = true
		}
		if entrySpan.IsError() {
			service.IsError = true
		}
	}
	if service.StartTime == 0 {
		service.StartTime = service.EntrySpans[0].StartTime
	}

	var trace *NodeSpanTrace = nil
	if sampledSpanTrace != nil || parent != nil {
		trace = newNodeSpanTrace(sampledSpanTrace, service)
		if parent == nil {
			traces.Traces = append(traces.Traces, trace)
		} else {
			parent.addChild(trace)
		}
	}
	for _, child := range service.Children {
		traces.collectApmTraceTree(apmType, child, sampledTraces, trace)
	}
}

type NodeSpanTrace struct {
	SampledTrace *model.Trace
	serviceNode  *apmmodel.OtelServiceNode
	children     []*NodeSpanTrace
	parent       *NodeSpanTrace
}

func newNodeSpanTrace(sampledTrace *model.Trace, serviceNode *apmmodel.OtelServiceNode) *NodeSpanTrace {
	return &NodeSpanTrace{
		SampledTrace: sampledTrace,
		serviceNode:  serviceNode,
		children:     make([]*NodeSpanTrace, 0),
	}
}

func (trace *NodeSpanTrace) addChild(child *NodeSpanTrace) {
	trace.children = append(trace.children, child)
	child.parent = trace
}

func (trace *NodeSpanTrace) GetClientCalls(spanId string) []*model.ApmClientCall {
	clientCalls := make([]*model.ApmClientCall, 0)
	node := trace.GetServiceNode(spanId)
	if node != nil {
		for _, child := range node.ExitSpans {
			nextEntryService := trace.GetServiceNode(child.NextSpanId)
			if nextEntryService == nil {
				clientCalls = append(clientCalls, NewApmClientCall(child, nil))
			} else {
				clientCalls = append(clientCalls, NewApmClientCall(child, nextEntryService.GetEntrySpan()))
			}
		}
	}
	return clientCalls
}

func (trace *NodeSpanTrace) GetServiceNode(spanId string) *apmmodel.OtelServiceNode {
	if trace.SampledTrace != nil && trace.SampledTrace.Labels.ApmSpanId == spanId {
		return trace.serviceNode
	}
	for _, child := range trace.children {
		if node := child.GetServiceNode(spanId); node != nil {
			return node
		}
	}
	return nil
}

func GetMatchSampledSpanTrace(apmType string, service *apmmodel.OtelServiceNode, sampledTraces map[string]*model.Trace) *model.Trace {
	for _, entrySpan := range service.EntrySpans {
		if apmType == "pinpoint" {
			for _, trace := range sampledTraces {
				if entrySpan.ServiceName == trace.Labels.ServiceName &&
					entrySpan.StartTime == trace.Labels.StartTime &&
					entrySpan.Duration == trace.Labels.Duration {
					return trace
				}
			}
		} else {
			if trace, ok := sampledTraces[entrySpan.SpanId]; ok {
				return trace
			}
		}
	}
	return nil
}
