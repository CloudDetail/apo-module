package client

import (
	"fmt"
	"math"

	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/CloudDetail/apo-module/model/v1"
)

func ConvertSlowTree(spanTrace *NodeSpanTrace) *TraceTree {
	traceTree := newTraceTree()
	traceTree.convertSlowTree(spanTrace, nil)

	return traceTree
}

func ConvertErrorTree(spanTrace *NodeSpanTrace) *ErrorTraceTree {
	traceTree := newErrorTraceTree()
	traceTree.convertErrorTree(spanTrace, nil)

	return traceTree
}

func BuildTopologyTree(trace *apmmodel.OTelTrace, sampledTraces *model.Traces) (*TraceTree, error) {
	spanTraces := make(map[string]*model.Trace, 0)
	mapSampleTraces(trace, sampledTraces)

	for _, sampledTrace := range sampledTraces.Traces {
		spanId := sampledTrace.Labels.ApmSpanId
		spanTraces[spanId] = sampledTrace

		service := trace.GetServiceNode(spanId)
		if service != nil {
			service.SetSpanId(spanId)

			if len(sampledTrace.Labels.Attributes) > 0 {
				service.Attribute = sampledTrace.Labels.Attributes
			}
		}
	}

	traceTree := newTraceTree()
	traceTree.collectApmTraceTree(trace.GetRoot(), nil, spanTraces)

	if traceTree.Root == nil {
		return nil, fmt.Errorf("no matched entry span is found in Apm System")
	}
	if !traceTree.Root.IsTraced {
		return nil, fmt.Errorf("entry[%s] is not collected by kindling", traceTree.Root.Id)
	}
	return traceTree, nil
}

func BuildErrorTree(trace *apmmodel.OTelTrace, sampledTraces *model.Traces) (*ErrorTraceTree, error) {
	spanTraces := make(map[string]*model.Trace, 0)
	mapSampleTraces(trace, sampledTraces)

	for _, sampledTrace := range sampledTraces.Traces {
		spanId := sampledTrace.Labels.ApmSpanId
		spanTraces[spanId] = sampledTrace
		service := trace.GetServiceNode(spanId)

		if service != nil {
			service.SetSpanId(spanId)
			if len(sampledTrace.Labels.Attributes) > 0 {
				service.Attribute = sampledTrace.Labels.Attributes
			}
		}
	}

	errorTree := newErrorTraceTree()
	errorTree.collectErrorTraceTree(trace.GetRoot(), nil, spanTraces)

	if errorTree.Root == nil {
		return nil, fmt.Errorf("no matched span is found in Apm System")
	}
	return errorTree, nil
}

func mapSampleTraces(trace *apmmodel.OTelTrace, sampledTraces *model.Traces) {
	if trace.ApmType != "pinpoint" {
		return
	}

	// For Pinpoint, use time range to match without spanId.
	for _, sampledTrace := range sampledTraces.Traces {
		if matchSpanId := getMatchSpanId(trace, sampledTrace.Labels); matchSpanId != "" {
			trace.MapSpanId(sampledTrace.Labels.ApmSpanId, matchSpanId)
		}
	}
}

func getMatchSpanId(trace *apmmodel.OTelTrace, sampledTrace *model.TraceLabels) string {
	matchSpanId := ""
	var matchDiff uint64 = math.MaxUint64
	for spanId, node := range trace.SpanServiceMap {
		for _, entrySpan := range node.EntrySpans {
			if entrySpan.ServiceName == sampledTrace.ServiceName {
				diff := getDiff(entrySpan, sampledTrace)
				if diff == 0 {
					// Whole Match
					return spanId
				}
				if matchDiff > diff {
					matchDiff = diff
					matchSpanId = spanId
				}
			}
		}
	}
	if matchDiff > 2000000 {
		// 1ms diff
		return ""
	}
	// Part Match
	return matchSpanId
}

/*
	   A         B      A       B    A       B           A    B
	   -----------      ---------    ---------           ------
	      -----     -------              --------      ----------
		  C   D     C     D              C      D      C        D
*/
func getDiff(entrySpan *apmmodel.OtelSpan, sampledTrace *model.TraceLabels) uint64 {
	if entrySpan.GetEndTime() < sampledTrace.StartTime {
		return entrySpan.Duration + sampledTrace.Duration
	}
	if entrySpan.StartTime > sampledTrace.EndTime {
		return entrySpan.Duration + sampledTrace.Duration
	}

	if entrySpan.StartTime == sampledTrace.StartTime && entrySpan.GetEndTime() == sampledTrace.EndTime {
		return 0
	}
	left := entrySpan.StartTime
	if left < sampledTrace.StartTime {
		left = sampledTrace.StartTime
	}
	right := entrySpan.GetEndTime()
	if right > sampledTrace.EndTime {
		right = sampledTrace.EndTime
	}
	return entrySpan.Duration + sampledTrace.Duration - 2*(right-left)
}

func GetClientCalls(trace *apmmodel.OTelTrace, spanId string) []*model.ApmClientCall {
	clientCalls := make([]*model.ApmClientCall, 0)
	serviceNode := trace.GetServiceNode(spanId)
	if serviceNode != nil {
		for _, child := range serviceNode.ExitSpans {
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

func NewApmClientCall(clientSpan *apmmodel.OtelSpan, serverEntrySpan *apmmodel.OtelSpan) *model.ApmClientCall {
	if serverEntrySpan == nil {
		return &model.ApmClientCall{
			ClientStartTime:      clientSpan.StartTime,
			ClientEndTime:        clientSpan.StartTime + clientSpan.Duration,
			ClientName:           clientSpan.Name,
			ClientSpanId:         clientSpan.SpanId,
			ClientAttributes:     clientSpan.Attributes,
			ClientOriginalSpanId: clientSpan.OriginalSpanId(),
			ServerDuration:       0,
		}
	}
	return &model.ApmClientCall{
		ClientStartTime:      clientSpan.StartTime,
		ClientEndTime:        clientSpan.StartTime + clientSpan.Duration,
		ClientName:           clientSpan.Name,
		ClientSpanId:         clientSpan.SpanId,
		ClientAttributes:     clientSpan.Attributes,
		ClientOriginalSpanId: clientSpan.OriginalSpanId(),
		ServerDuration:       serverEntrySpan.Duration,
		ServerName:           serverEntrySpan.ServiceName,
	}
}

func GetErrorSpans(node *apmmodel.OtelServiceNode) []*model.ErrorSpan {
	errorSpans := make([]*model.ErrorSpan, 0)

	for _, errorSpan := range node.ErrorSpans {
		errorSpans = append(errorSpans, convertErrorSpan(errorSpan))
	}
	for _, entrySpan := range node.EntrySpans {
		if len(entrySpan.Exceptions) > 0 {
			errorSpans = append(errorSpans, convertErrorSpan(entrySpan))
		}
	}

	var (
		errStatusCode string
		errEntrySpan  *apmmodel.OtelSpan
	)
	for _, entrySpan := range node.EntrySpans {
		if entrySpan.IsError() && len(entrySpan.Exceptions) == 0 {
			if statusCode, exist := entrySpan.Attributes[apmmodel.AttributeHTTPStatusCode]; exist {
				errStatusCode = statusCode
				errEntrySpan = entrySpan
			}
		}
	}
	if errStatusCode != "" && errStatusCode != "200" {
		apmErrorSpan := model.NewErrorSpan(
			errEntrySpan.Name,
			errEntrySpan.StartTime,
			errEntrySpan.Duration,
		)
		apmErrorSpan.Exceptions = append(apmErrorSpan.Exceptions, model.NewOtelException(
			(errEntrySpan.StartTime+errEntrySpan.Duration)/1000, // ns -> us
			"HTTP ERROR CODE",
			fmt.Sprintf("HTTP ERROR CODE: %s", errStatusCode),
			"",
		))
		errorSpans = append(errorSpans, apmErrorSpan)
	}
	return errorSpans
}

func convertErrorSpan(span *apmmodel.OtelSpan) *model.ErrorSpan {
	apmErrorSpan := model.NewErrorSpan(
		span.Name,
		span.StartTime,
		span.Duration,
	)
	for k, v := range span.Attributes {
		apmErrorSpan.AddAttribute(k, v)
	}

	apmErrorSpan.Exceptions = append(apmErrorSpan.Exceptions, span.Exceptions...)
	return apmErrorSpan
}
