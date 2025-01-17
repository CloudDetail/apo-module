package model

import (
	"fmt"
)

type TraceTreeNode struct {
	Id                string         `json:"id"`
	ServiceName       string         `json:"serviceName"`
	Url               string         `json:"url"`
	StartTime         uint64         `json:"startTime"`
	TotalTime         uint64         `json:"totalTime"`
	ClientTime        uint64         `json:"clientTime"`
	P90               uint64         `json:"p90"`
	ThresholdType     ThresholdType  `json:"threshold_type"`
	ThresholdValue    float64        `json:"threshold_value"`
	ThresholdRange    ThresholdRange `json:"threshold_range"`
	ThresholdMultiple float64        `json:"threshold_multiple"`

	IsTraced       bool             `json:"isTraced"`
	IsProfiled     bool             `json:"isProfiled"`
	Pod            string           `json:"pod"`
	PodNS          string           `json:"podNS"`
	Workload       string           `json:"workload"`
	WorkloadType   string           `json:"workloadType"`
	IsPath         bool             `json:"isPath"`
	IsMutated      bool             `json:"isMutated"`
	MissVNode      bool             `json:"missVNode"`
	SelfTime       uint64           `json:"selfTime"`
	SelfP90        uint64           `json:"selfP90"`
	MutatedValue   int64            `json:"mutatedValue"`
	SpanId         string           `json:"spanId"`
	OriginalSpanId string           `json:"-"`
	ContainerId    string           `json:"-"`
	NodeIp         string           `json:"-"`
	NodeName       string           `json:"-"`
	Pid            uint32           `json:"-"`
	Children       []*TraceTreeNode `json:"children"`
	Parent         *TraceTreeNode   `json:"-"`
}

func (node *TraceTreeNode) GetMutatedSpanId() string {
	if node.IsMutated {
		return node.SpanId
	}
	for _, child := range node.Children {
		if child.IsPath {
			return child.GetMutatedSpanId()
		}
	}
	return ""
}

func (node *TraceTreeNode) HasVNodeChild() bool {
	for _, child := range node.Children {
		if child.MissVNode {
			return true
		}
	}
	return false
}

func (node *TraceTreeNode) SetSampled(sampledTrace *Trace) {
	node.Id = sampledTrace.GetInstanceId()
	sampledTraceLabel := sampledTrace.Labels
	node.Url = sampledTraceLabel.Url
	node.P90 = uint64(sampledTraceLabel.ThresholdValue / sampledTraceLabel.ThresholdMultiple)
	node.ThresholdValue = sampledTraceLabel.ThresholdValue
	node.ThresholdType = sampledTraceLabel.ThresholdType
	node.ThresholdRange = sampledTraceLabel.ThresholdRange
	node.ThresholdMultiple = sampledTraceLabel.ThresholdMultiple
	node.IsTraced = true
	node.IsProfiled = sampledTraceLabel.IsProfiled
	node.Pid = sampledTraceLabel.Pid
	node.ContainerId = sampledTraceLabel.ContainerId
	node.NodeIp = sampledTraceLabel.NodeIp
	node.NodeName = sampledTraceLabel.NodeName
	node.Pod = sampledTrace.PodName
	node.PodNS = sampledTrace.Namespace
	node.Workload = sampledTrace.WorkloadName
	node.WorkloadType = sampledTrace.WorkloadKind
}

func (node *TraceTreeNode) AddChild(child *TraceTreeNode) {
	node.Children = append(node.Children, child)
	child.Parent = node
}

func (node *TraceTreeNode) CheckP90() error {
	if node.P90 == 0 {
		return fmt.Errorf("p90 is not found for Instance(%s)", node.Id)
	}
	for _, child := range node.Children {
		if err := child.CheckP90(); err != nil {
			return err
		}
	}
	return nil
}

func (node *TraceTreeNode) CalcMutateValue() int64 {
	if node.MutatedValue == 0 {
		var outTime uint64 = 0
		for _, child := range node.Children {
			outTime += child.TotalTime
		}
		if node.TotalTime > outTime {
			node.SelfTime = node.TotalTime - outTime
		} else {
			node.SelfTime = 0
		}

		var outP90 uint64 = 0
		for _, child := range node.Children {
			if child.IsTraced {
				outP90 += child.P90
			} else {
				outP90 += child.TotalTime
			}
		}
		if node.P90 > 0 {
			if node.P90 >= outP90 {
				node.SelfP90 = node.P90 - outP90
			} else {
				node.SelfP90 = node.P90 / 2
			}
			node.MutatedValue = int64(node.SelfTime) - int64(node.SelfP90)
		} else {
			node.SelfP90 = 0
		}
	}
	return node.MutatedValue
}

func (node *TraceTreeNode) MarkPath() {
	node.IsPath = true
	if node.Parent != nil {
		node.Parent.MarkPath()
	}
}
