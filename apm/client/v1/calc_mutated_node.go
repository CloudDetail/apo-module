package client

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/CloudDetail/apo-module/model/v1"
)

type CalcMutatedNodeFn func(tree *TraceTree, traceId string, ratioThreshold int) (*model.TraceTreeNode, error)

var CalcByMutatedSpan CalcMutatedNodeFn = func(tree *TraceTree, traceId string, ratioThreshold int) (*model.TraceTreeNode, error) {
	sortedNodes := make([]*model.TraceTreeNode, 0)
	for _, v := range tree.NodeMap {
		v.CalcMutateValue()
		if tree.Root.SpanId == v.SpanId && v.HasVNodeChild() {
			continue
		}
		sortedNodes = append(sortedNodes, v)
	}
	sort.Sort(byMuatedValue(sortedNodes))

	duration := tree.Root.TotalTime
	threshold := duration * uint64(ratioThreshold) / 100
	var profiledNode *model.TraceTreeNode = nil
	node := sortedNodes[0]
	if node.MutatedValue > 0 {
		if node.SelfTime >= threshold {
			return node, nil
		}
		if profiledNode == nil || profiledNode.SelfTime < node.SelfTime {
			profiledNode = node
		}
	}

	if profiledNode == nil {
		return nil, fmt.Errorf("Instance(%s) is not mutated. Mutated[%d], Self: %d", node.Id, node.MutatedValue, node.SelfTime)
	}

	var percent float64 = 0
	if duration > 0 {
		percent = float64(profiledNode.SelfTime*100) / float64(duration)
	}
	return nil, fmt.Errorf("Instance(%s) selfTime(%sms) has not enough duration ratio(%s%%)",
		profiledNode.Id,
		strconv.FormatFloat(float64(profiledNode.SelfTime)/1000000, 'f', 2, 64),
		strconv.FormatFloat(percent, 'f', 2, 64))
}

var CalcByMutatedService CalcMutatedNodeFn = func(tree *TraceTree, traceId string, ratioThreshold int) (*model.TraceTreeNode, error) {
	serviceEndPoints := newServiceEndPoints()
	for _, v := range tree.NodeMap {
		v.CalcMutateValue()
		serviceEndPoints.addMutateSpan(v)
	}
	sort.Sort(byServiceMutatedValue(serviceEndPoints.services))

	if len(serviceEndPoints.services) == 0 {
		return nil, fmt.Errorf("trace[%s] has no mutated service", traceId)
	}
	serviceEndPoint := serviceEndPoints.services[0]
	duration := tree.Root.TotalTime
	threshold := duration * uint64(ratioThreshold) / 100
	if serviceEndPoint.SelfTime < threshold {
		var percent float64 = 0
		if duration > 0 {
			percent = float64(serviceEndPoint.SelfTime*100) / float64(duration)
		}
		return nil, fmt.Errorf("service(%s) selfTime(%sms) has not enough duration ratio(%s%%)",
			serviceEndPoint.ServiceName,
			strconv.FormatFloat(float64(serviceEndPoint.SelfTime)/1000000, 'f', 2, 64),
			strconv.FormatFloat(percent, 'f', 2, 64))
	}

	node, _ := serviceEndPoint.getMutateNode()
	if node.MutatedValue > 0 {
		return node, nil
	}
	return nil, fmt.Errorf("Instance(%s) URL(%s) is not mutated. Mutated[%d], Self: %d", node.Id, node.Url, node.MutatedValue, node.SelfTime)
}

var CalcByTop3MutatedService CalcMutatedNodeFn = func(tree *TraceTree, traceId string, ratioThreshold int) (*model.TraceTreeNode, error) {
	serviceEndPoints := newServiceEndPoints()
	for _, v := range tree.NodeMap {
		v.CalcMutateValue()
		serviceEndPoints.addMutateSpan(v)
	}
	sort.Sort(byServiceMutatedValue(serviceEndPoints.services))

	if len(serviceEndPoints.services) == 0 {
		return nil, fmt.Errorf("trace[%s] has no mutated service", traceId)
	}
	duration := tree.Root.TotalTime
	threshold := duration * uint64(ratioThreshold) / 100
	var profiledMutatedNode *model.TraceTreeNode
	var profiled bool
	for i, serviceEndPoint := range serviceEndPoints.services {
		if i == 3 {
			break
		}
		if serviceEndPoint.SelfTime < threshold {
			var percent float64 = 0
			if duration > 0 {
				percent = float64(serviceEndPoint.SelfTime*100) / float64(duration)
			}
			log.Printf("The Top[%d] service(%s) selfTime(%sms) has not enough duration ratio(%s%%)",
				i+1, serviceEndPoint.ServiceName,
				strconv.FormatFloat(float64(serviceEndPoint.SelfTime)/1000000, 'f', 2, 64),
				strconv.FormatFloat(percent, 'f', 2, 64))
			continue
		}

		profiledMutatedNode, profiled = serviceEndPoint.getMutateNode()
		if profiledMutatedNode.MutatedValue > 0 && profiled {
			return profiledMutatedNode, nil
		}
	}

	if profiledMutatedNode == nil {
		return nil, fmt.Errorf("no Top3 service has enough duration ratio")
	} else {
		return nil, fmt.Errorf("top3 node [%s] has enough duration ratio but is not profiled",
			profiledMutatedNode.Id)
	}
}
