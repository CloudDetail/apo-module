package model

import (
	"fmt"
)

type OTelTrace struct {
	SpanServiceMap  map[string]*OtelServiceNode
	rootServiceNode *OtelServiceNode
	spanIdMap       map[string]string
	ApmType         string
}

func NewOTelTrace(apmType string) *OTelTrace {
	return &OTelTrace{
		SpanServiceMap:  make(map[string]*OtelServiceNode, 0),
		rootServiceNode: nil,
		spanIdMap:       make(map[string]string, 0),
		ApmType:         apmType,
	}
}

func (trace *OTelTrace) MapSpanId(realId string, spanId string) {
	trace.spanIdMap[realId] = spanId
}

func (trace *OTelTrace) AddServiceNode(serviceNode *OtelServiceNode, parent *OtelServiceNode) error {
	for _, entrySpan := range serviceNode.EntrySpans {
		trace.SpanServiceMap[entrySpan.SpanId] = serviceNode
		if entrySpan.PSpanId == "" {
			serviceNode.IsRoot = true
		}
		if entrySpan.IsError() {
			serviceNode.IsError = true
		}
		serviceNode.Parent = parent
	}

	for _, child := range serviceNode.Children {
		trace.AddServiceNode(child, serviceNode)
	}
	return nil
}

func (trace *OTelTrace) GetServiceNode(spanId string) *OtelServiceNode {
	if node, ok := trace.SpanServiceMap[spanId]; ok {
		return node
	}
	if matchId, exist := trace.spanIdMap[spanId]; exist {
		return trace.SpanServiceMap[matchId]
	}
	return nil
}

func (trace *OTelTrace) GetServiceNodes() []*OtelServiceNode {
	nodes := make([]*OtelServiceNode, 0)

	if trace.rootServiceNode == nil {
		for _, node := range trace.SpanServiceMap {
			if node.Parent == nil {
				nodes = append(nodes, node)
			}
		}
	} else {
		nodes = append(nodes, trace.rootServiceNode)
	}
	return nodes
}

func (trace *OTelTrace) addServiceNode(spanId string, serviceNode *OtelServiceNode) error {
	if serviceNode.IsRoot {
		if trace.rootServiceNode == nil {
			trace.rootServiceNode = serviceNode
		} else {
			return fmt.Errorf("more than one Root Entry: %s, %s", trace.rootServiceNode.ServiceName, serviceNode.ServiceName)
		}
	}
	trace.SpanServiceMap[spanId] = serviceNode
	return nil
}

func (trace *OTelTrace) cacheSpanIdServiceNode(spanId string, serviceNode *OtelServiceNode) {
	trace.SpanServiceMap[spanId] = serviceNode
}

func (trace *OTelTrace) RelateServices(service *OtelServiceNode, spanId string, serviceNodes map[string]*OtelServiceNode, spanMap map[string]*OtelSpan, childSpansMap map[string][]string) {
	trace.addServiceNode(spanId, service)

	var exitSpan *OtelSpan
	span := spanMap[spanId]
	if span.Kind.IsExit() {
		service.AddExitSpan(span)
		exitSpan = span
	} else if span.Kind.IsEntry() {
		service.addEntrySpan(span)
	}

	if children, exist := childSpansMap[spanId]; exist {
		for _, child := range children {
			if childService, found := serviceNodes[child]; found {
				service.addChild(childService)
				if exitSpan != nil {
					exitSpan.NextSpanId = childService.EntrySpans[0].SpanId
				}
			} else {
				trace.RelateServices(service, child, serviceNodes, spanMap, childSpansMap)
			}
		}
	}
}

func (trace *OTelTrace) GetRoot() *OtelServiceNode {
	return trace.rootServiceNode
}

func (trace *OTelTrace) SetRoot(root *OtelServiceNode) {
	trace.rootServiceNode = root
}

func (trace *OTelTrace) CheckRoot(serviceNodes []*OtelServiceNode) error {
	for _, serviceNode := range serviceNodes {
		if serviceNode.IsRoot {
			if trace.rootServiceNode == nil {
				trace.rootServiceNode = serviceNode
			} else {
				return fmt.Errorf("more than one Root Entry: %s, %s", trace.rootServiceNode.EntrySpans[0].ServiceName, serviceNode.EntrySpans[0].ServiceName)
			}
		}
	}
	return nil
}
