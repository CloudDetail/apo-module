package client

import (
	"fmt"
	"sort"

	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/CloudDetail/apo-module/model/v1"
)

type ErrorTraceTree struct {
	Root    *model.ErrorTreeNode
	NodeMap map[string]*model.ErrorTreeNode
}

func newErrorTraceTree() *ErrorTraceTree {
	return &ErrorTraceTree{
		Root:    nil,
		NodeMap: make(map[string]*model.ErrorTreeNode, 0),
	}
}

func (tree *ErrorTraceTree) addTraceNode(parent *model.ErrorTreeNode, child *model.ErrorTreeNode) *model.ErrorTreeNode {
	if tree.Root == nil {
		tree.Root = child
		child.Depth = 1
	} else {
		child.Depth = parent.Depth + 1
		parent.AddChild(child)
	}
	tree.NodeMap[child.SpanId] = child
	return child
}

func (tree *ErrorTraceTree) collectErrorTraceTree(node *apmmodel.OtelServiceNode, parentNode *model.ErrorTreeNode, sampledNodeMap map[string]*model.Trace) {
	currentNode := tree.addTraceNode(parentNode, newApmErrorTreeNode(node))
	currentNode.ErrorSpans = GetErrorSpans(node)

	if sampledTrace, exist := sampledNodeMap[node.SpanId]; exist {
		currentNode.SetSampled(sampledTrace)
	}
	for _, child := range node.Children {
		tree.collectErrorTraceTree(child, currentNode, sampledNodeMap)
	}
}

func (tree *ErrorTraceTree) convertErrorTree(trace *NodeSpanTrace, parentNode *model.ErrorTreeNode) {
	currentNode := tree.addTraceNode(parentNode, newApmErrorTreeNode(trace.serviceNode))
	currentNode.ErrorSpans = GetErrorSpans(trace.serviceNode)

	if trace.SampledTrace != nil {
		currentNode.SetSampled(trace.SampledTrace)
	}
	for _, child := range trace.children {
		tree.convertErrorTree(child, currentNode)
	}
}

func (tree *ErrorTraceTree) GetRootCauseErrorNode(traceId string) (*model.ErrorTreeNode, error) {
	nodes := make([]*model.ErrorTreeNode, 0)
	for _, v := range tree.NodeMap {
		if v.IsError {
			nodes = append(nodes, v)
		}
	}
	sort.Sort(model.ByErrorDepth(nodes))

	if len(nodes) == 0 {
		return nil, fmt.Errorf("trace[%s] has no traced trace", traceId)
	}
	errorNode := nodes[0]
	if !errorNode.IsTraced {
		return nil, fmt.Errorf("trace[%s] span(%s) is not traced", traceId, errorNode.SpanId)
	}
	errorNode.IsMutated = true
	errorNode.MarkPath()
	return errorNode, nil
}

func newApmErrorTreeNode(node *apmmodel.OtelServiceNode) *model.ErrorTreeNode {
	entrySpan := node.GetEntrySpan()
	return &model.ErrorTreeNode{
		Id:             entrySpan.ServiceName,
		ServiceName:    entrySpan.ServiceName,
		Url:            entrySpan.Name,
		StartTime:      entrySpan.StartTime,
		TotalTime:      entrySpan.Duration,
		IsTraced:       false,
		IsProfiled:     false,
		IsSampled:      !entrySpan.NotSampled,
		IsError:        node.IsError,
		IsMutated:      false,
		MissVNode:      node.VNode,
		SpanId:         node.SpanId,
		OriginalSpanId: node.OriginalSpanId,
		Depth:          0,
		Children:       make([]*model.ErrorTreeNode, 0),
		ErrorSpans:     make([]*model.ErrorSpan, 0),
	}
}
