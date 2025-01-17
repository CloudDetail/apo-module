package model

type ErrorTreeNode struct {
	Id                string         `json:"id"`
	ServiceName       string         `json:"serviceName"`
	Url               string         `json:"url"`
	StartTime         uint64         `json:"startTime"`
	TotalTime         uint64         `json:"totalTime"`
	P90               uint64         `json:"p90"`
	ThresholdType     ThresholdType  `json:"threshold_type"`
	ThresholdValue    float64        `json:"threshold_value"`
	ThresholdRange    ThresholdRange `json:"threshold_range"`
	ThresholdMultiple float64        `json:"threshold_multiple"`
	IsSampled         bool           `json:"isSampled"`

	IsTraced       bool             `json:"isTraced"`
	IsProfiled     bool             `json:"isProfiled"`
	Pod            string           `json:"pod"`
	PodNS          string           `json:"podNS"`
	Workload       string           `json:"workload"`
	WorkloadType   string           `json:"workloadType"`
	IsError        bool             `json:"isError"`
	IsPath         bool             `json:"isPath"`
	IsMutated      bool             `json:"isMutated"`
	MissVNode      bool             `json:"missVNode"`
	SpanId         string           `json:"spanId"`
	OriginalSpanId string           `json:"-"`
	Depth          int              `json:"depth"`
	ContainerId    string           `json:"-"`
	NodeIp         string           `json:"-"`
	NodeName       string           `json:"nodeName"`
	Pid            uint32           `json:"-"`
	Children       []*ErrorTreeNode `json:"children"`
	ErrorSpans     []*ErrorSpan     `json:"errorSpans"`
	Parent         *ErrorTreeNode   `json:"-"`
}

func (node *ErrorTreeNode) GetRootCauseError() *Exception {
	var earliestException *Exception

	for _, errorSpan := range node.ErrorSpans {
		for _, exception := range errorSpan.Exceptions {
			if earliestException == nil || exception.Timestamp < earliestException.Timestamp {
				earliestException = exception
			}
		}
	}

	return earliestException
}

func (node *ErrorTreeNode) SetSampled(sampledTrace *Trace) {
	node.Id = sampledTrace.GetInstanceId()
	sampledTraceLabel := sampledTrace.Labels
	node.Url = sampledTraceLabel.Url
	node.P90 = uint64(sampledTraceLabel.ThresholdValue / sampledTraceLabel.ThresholdMultiple)
	node.ThresholdValue = sampledTraceLabel.ThresholdValue
	node.ThresholdType = sampledTraceLabel.ThresholdType
	node.ThresholdRange = sampledTraceLabel.ThresholdRange
	node.ThresholdMultiple = sampledTraceLabel.ThresholdMultiple
	node.IsError = sampledTraceLabel.IsError
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

func (node *ErrorTreeNode) AddChild(child *ErrorTreeNode) {
	node.Children = append(node.Children, child)
	child.Parent = node
}

func (node *ErrorTreeNode) GetErrorDepth() int {
	if !node.IsError {
		return 0
	}
	return node.Depth
}

func (node *ErrorTreeNode) MarkPath() {
	node.IsPath = true
	if node.Parent != nil {
		node.Parent.MarkPath()
	}
}

type ByErrorDepth []*ErrorTreeNode

func (m ByErrorDepth) Len() int {
	return len(m)
}

func (m ByErrorDepth) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m ByErrorDepth) Less(i, j int) bool {
	return m[i].GetErrorDepth() > m[j].GetErrorDepth()
}
