package model

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var ExceptionAsError = strings.ToUpper(os.Getenv("EXCEPTION_AS_ERROR")) == "TRUE"

type OtelTree struct {
	SpanMap  map[string]*OtelSpan
	Children map[string][]string
	rootSpan *OtelSpan
}

func NewOtelTree() *OtelTree {
	return &OtelTree{
		SpanMap:  make(map[string]*OtelSpan),
		Children: make(map[string][]string),
	}
}

func (tree *OtelTree) AddSpan(span *OtelSpan) error {
	if span.PSpanId == "" {
		if tree.rootSpan != nil {
			return fmt.Errorf("more than one Root Span: %s, %s", tree.rootSpan.SpanId, span.SpanId)
		} else {
			tree.rootSpan = span
		}
	}
	if _, exist := tree.SpanMap[span.SpanId]; exist {
		log.Printf("Ignore Repeated Span %s\n", span.SpanId)
		return nil
	}
	tree.SpanMap[span.SpanId] = span

	if span.PSpanId != "" {
		childSpans, found := tree.Children[span.PSpanId]
		if !found {
			childSpans = make([]string, 0)
		}
		childSpans = append(childSpans, span.SpanId)
		tree.Children[span.PSpanId] = childSpans
	}
	return nil
}

func (tree *OtelTree) GetRoot() *OtelSpan {
	return tree.rootSpan
}

func (tree *OtelTree) BuildServiceNodes(trace *OTelTrace) error {
	for _, span := range tree.SpanMap {
		if span.Kind.IsEntry() {
			serviceNode := newOTelServiceNode(span)
			if err := trace.addServiceNode(span.SpanId, serviceNode); err != nil {
				return err
			}
		}
	}
	for pspanId, ChildrenIds := range tree.Children {
		if parentSpan, exist := tree.SpanMap[pspanId]; exist {
			serviceNode := trace.GetServiceNode(parentSpan.SpanId)
			for _, childSpanId := range ChildrenIds {
				childSpan := tree.SpanMap[childSpanId]
				if childSpan.Kind.IsEntry() {
					serviceNode.addChild(trace.GetServiceNode(childSpanId))
				} else {
					if childSpan.Kind.IsExit() {
						serviceNode.AddExitSpan(childSpan)
					} else {
						serviceNode.addEntrySpan(childSpan)
						trace.cacheSpanIdServiceNode(childSpan.SpanId, serviceNode)
					}
				}
			}
		}
	}
	return nil
}

func (tree *OtelTree) BuildRelation4Spans(trace *OTelTrace) error {
	var rootServiceNode *OtelServiceNode = nil
	serviceNodes := make(map[string]*OtelServiceNode, 0)
	missParentEntrySpanIds := make([]string, 0)
	for spanId, span := range tree.SpanMap {
		parentSpanId := span.PSpanId
		if len(parentSpanId) > 0 {
			if parentSpan, exist := tree.SpanMap[parentSpanId]; exist {
				if parentSpan.Kind.IsExit() && tree.isChildEntry(span) {
					serviceNodes[spanId] = newOTelServiceNode(span)
				}
			} else if span.Kind.IsEntry() {
				serviceNodes[spanId] = newOTelServiceNode(span)
				missParentEntrySpanIds = append(missParentEntrySpanIds, spanId)
			}
		} else {
			if rootServiceNode == nil {
				rootServiceNode = newOTelServiceNode(span)
				serviceNodes[spanId] = rootServiceNode
			} else {
				return fmt.Errorf("multi RootSpans: %s and %s", rootServiceNode.ServiceName, span.ServiceName)
			}
		}
	}

	for spanId, service := range serviceNodes {
		trace.RelateServices(service, spanId, serviceNodes, tree.SpanMap, tree.Children)
		service.RelateExceptions(spanId, tree.SpanMap, tree.Children)
		if ExceptionAsError && service.HasException {
			service.IsError = true
		}
		if service.IsError {
			service.RelateErrors(spanId, tree.SpanMap, tree.Children)
		}
	}
	trace.rootServiceNode = rootServiceNode
	trace.missParentEntrySpanIds = missParentEntrySpanIds
	return nil
}

func (tree *OtelTree) isChildEntry(span *OtelSpan) bool {
	if span.Kind.IsEntry() {
		return true
	}
	if span.Kind.IsExit() {
		return false
	}
	if childSpanIds, found := tree.Children[span.SpanId]; found {
		for _, childSpanId := range childSpanIds {
			if childSpan, exist := tree.SpanMap[childSpanId]; exist {
				if tree.isChildEntry(childSpan) {
					return true
				}
			}
		}
	}
	return false
}

func (tree *OtelTree) ConvertToService() (*OtelServiceNode, error) {
	var rootServiceNode *OtelServiceNode = nil
	serviceNodes := make(map[string]*OtelServiceNode, 0)
	for spanId, span := range tree.SpanMap {
		if len(span.PSpanId) > 0 {
			if parentSpan, exist := tree.SpanMap[span.PSpanId]; exist {
				if parentSpan.Kind.IsExit() {
					serviceNodes[spanId] = newOTelServiceNode(span)
				}
			} else {
				return nil, fmt.Errorf("miss ParentSpan for SpanTree: %s, span: %s, service: %s", span.PSpanId, span.Name, span.ServiceName)
			}
		} else {
			if rootServiceNode == nil {
				rootServiceNode = newOTelServiceNode(span)
				serviceNodes[spanId] = rootServiceNode
			} else {
				return nil, fmt.Errorf("multi RootSpans: %s and %s", rootServiceNode.ServiceName, span.ServiceName)
			}
		}
	}

	if rootServiceNode == nil {
		return nil, fmt.Errorf("miss RootSpan")
	}
	return nil, nil
}
