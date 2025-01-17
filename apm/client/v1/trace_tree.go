package client

import (
	"github.com/CloudDetail/apo-module/model/v1"

	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
)

type TraceTree struct {
	Root    *model.TraceTreeNode
	NodeMap map[string]*model.TraceTreeNode
}

func newTraceTree() *TraceTree {
	return &TraceTree{
		Root:    nil,
		NodeMap: make(map[string]*model.TraceTreeNode, 0),
	}
}

func (tree *TraceTree) addTraceNode(parent *model.TraceTreeNode, child *model.TraceTreeNode) *model.TraceTreeNode {
	if tree.Root == nil {
		tree.Root = child
	} else {
		parent.AddChild(child)
	}
	tree.NodeMap[child.SpanId] = child
	return child
}

func (tree *TraceTree) collectApmTraceTree(node *apmmodel.OtelServiceNode, parentNode *model.TraceTreeNode, sampledNodeMap map[string]*model.Trace) {
	currentNode := tree.addTraceNode(parentNode, newApmTraceTreeNode(node))

	if sampledTrace, exist := sampledNodeMap[node.SpanId]; exist {
		currentNode.SetSampled(sampledTrace)
	}
	for _, child := range node.Children {
		tree.collectApmTraceTree(child, currentNode, sampledNodeMap)
	}
}

func (tree *TraceTree) convertSlowTree(trace *NodeSpanTrace, parentNode *model.TraceTreeNode) {
	currentNode := tree.addTraceNode(parentNode, newApmTraceTreeNode(trace.serviceNode))

	if trace.SampledTrace != nil {
		currentNode.SetSampled(trace.SampledTrace)
	}
	for _, child := range trace.children {
		tree.convertSlowTree(child, currentNode)
	}
}

func (tree *TraceTree) GetMutatedTraceNode(traceId string, ratioThreshold int, mode string) (*model.TraceTreeNode, error) {
	mutatedNode, err := tree.calcMutatedNode(traceId, ratioThreshold, mode)
	if err != nil {
		return nil, err
	}

	mutatedNode.IsMutated = true
	if mutatedTrace, exist := tree.NodeMap[mutatedNode.SpanId]; exist {
		mutatedTrace.MarkPath()
	}
	return mutatedNode, nil
}

func (tree *TraceTree) calcMutatedNode(traceId string, ratioThreshold int, mode string) (*model.TraceTreeNode, error) {
	if mode == "single" {
		return CalcByMutatedSpan(tree, traceId, ratioThreshold)
	} else if mode == "maxService" {
		return CalcByMutatedService(tree, traceId, ratioThreshold)
	}
	return CalcByTop3MutatedService(tree, traceId, ratioThreshold)
}

func newApmTraceTreeNode(node *apmmodel.OtelServiceNode) *model.TraceTreeNode {
	var clientTime uint64 = 0
	if clientSpan := node.GetClientSpan(); clientSpan != nil {
		clientTime = clientSpan.Duration
	}
	entrySpan := node.GetEntrySpan()
	return &model.TraceTreeNode{
		Id:             entrySpan.ServiceName,
		ServiceName:    entrySpan.ServiceName,
		Url:            entrySpan.Name,
		StartTime:      entrySpan.StartTime,
		TotalTime:      entrySpan.Duration,
		ClientTime:     clientTime,
		P90:            0,
		IsTraced:       false,
		IsProfiled:     false,
		IsPath:         false,
		IsMutated:      false,
		MissVNode:      node.VNode,
		SelfTime:       0,
		SelfP90:        0,
		MutatedValue:   0,
		SpanId:         node.SpanId,
		OriginalSpanId: node.OriginalSpanId,
		Children:       make([]*model.TraceTreeNode, 0),
	}
}
