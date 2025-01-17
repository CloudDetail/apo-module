package model

type OtelServiceNode struct {
	StartTime      uint64             `json:"-"`
	ServiceName    string             `json:"-"`
	EntrySpans     []*OtelSpan        `json:"entrySpans"`
	ExitSpans      []*OtelSpan        `json:"exitSpans,omitempty"`
	ErrorSpans     []*OtelSpan        `json:"errorSpans,omitempty"`
	Children       []*OtelServiceNode `json:"children,omitempty"`
	Parent         *OtelServiceNode   `json:"-"`
	VNode          bool               `json:"-"`
	SpanId         string             `json:"-"`
	OriginalSpanId string             `json:"-"`
	IsRoot         bool               `json:"-"`
	IsError        bool               `json:"-"`
	HasException   bool               `json:"-"`
	Attribute      string             `json:"-"`
}

func newOTelServiceNode(span *OtelSpan) *OtelServiceNode {
	return &OtelServiceNode{
		ServiceName: span.ServiceName,
		EntrySpans:  []*OtelSpan{span},
		IsRoot:      span.PSpanId == "",
		IsError:     span.IsError(),
	}
}

func (serviceNode *OtelServiceNode) SetSpanId(spanId string) {
	serviceNode.SpanId = spanId
	for _, entrySpan := range serviceNode.EntrySpans {
		if entrySpan.SpanId == spanId {
			serviceNode.OriginalSpanId = entrySpan.OriginalSpanId()
		}
	}
}

func (serviceNode *OtelServiceNode) addEntrySpan(span *OtelSpan) {
	for _, entrySpan := range serviceNode.EntrySpans {
		if entrySpan.SpanId == span.SpanId {
			return
		}
	}
	serviceNode.EntrySpans = append(serviceNode.EntrySpans, span)
	if span.PSpanId == "" {
		serviceNode.IsRoot = true
	}
	if span.IsError() {
		serviceNode.IsError = true
	}
}

func (serviceNode *OtelServiceNode) addChild(child *OtelServiceNode) {
	serviceNode.Children = append(serviceNode.Children, child)
	child.Parent = serviceNode
}

func (serviceNode *OtelServiceNode) AddExitSpan(exitSpan *OtelSpan) {
	serviceNode.ExitSpans = append(serviceNode.ExitSpans, exitSpan)
}

func (serviceNode *OtelServiceNode) AddErrorSpan(errorSpan *OtelSpan) {
	serviceNode.ErrorSpans = append(serviceNode.ErrorSpans, errorSpan)
}

func (serviceNode *OtelServiceNode) RelateExceptions(spanId string, spanMap map[string]*OtelSpan, childSpansMap map[string][]string) {
	if span, ok := spanMap[spanId]; ok {
		if len(span.Exceptions) > 0 {
			serviceNode.HasException = true
		}
		if span.Kind.IsExit() {
			return
		}
		if childSpanIds, exist := childSpansMap[spanId]; exist {
			for _, childSpanId := range childSpanIds {
				serviceNode.RelateExceptions(childSpanId, spanMap, childSpansMap)
			}
		}
	}
}

func (serviceNode *OtelServiceNode) RelateErrors(spanId string, spanMap map[string]*OtelSpan, childSpansMap map[string][]string) {
	if span, ok := spanMap[spanId]; ok {
		if span.IsError() && !span.Kind.IsEntry() {
			serviceNode.AddErrorSpan(span)
		}
		if span.Kind.IsExit() {
			return
		}
		if childSpanIds, exist := childSpansMap[spanId]; exist {
			for _, childSpanId := range childSpanIds {
				serviceNode.RelateErrors(childSpanId, spanMap, childSpansMap)
			}
		}
	}
}

func (serviceNode *OtelServiceNode) GetStartTime() uint64 {
	return serviceNode.EntrySpans[0].StartTime
}

func (node *OtelServiceNode) SetFixTime() {
	if node.StartTime == 0 {
		node.StartTime = node.EntrySpans[0].StartTime
	}
	clientSpanMap := make(map[string]*OtelSpan)
	for _, exitSpan := range node.ExitSpans {
		if exitSpan.NextSpanId != "" {
			clientSpanMap[exitSpan.NextSpanId] = exitSpan
		}
	}

	for _, child := range node.Children {
		matchClient := clientSpanMap[child.EntrySpans[0].SpanId]
		if child.StartTime == 0 && matchClient != nil && matchClient.Duration > child.EntrySpans[0].StartTime {
			child.StartTime = node.StartTime + matchClient.GetEndTime() - child.EntrySpans[0].GetEndTime()
		}
		child.SetFixTime()
	}
}

func (serviceNode *OtelServiceNode) CheckVNode(parentSpanId string) {
	for _, entrySpan := range serviceNode.EntrySpans {
		if entrySpan.PSpanId == parentSpanId {
			return
		}
	}
	serviceNode.VNode = true
}

func (serviceNode *OtelServiceNode) MatchEntrySpan(spanId string) bool {
	for _, entrySpan := range serviceNode.EntrySpans {
		if entrySpan.SpanId == spanId {
			return true
		}
	}
	return false
}

func (serviceNode *OtelServiceNode) GetClientSpan() *OtelSpan {
	if serviceNode.Parent == nil {
		return nil
	}
	for _, exitSpan := range serviceNode.Parent.ExitSpans {
		if serviceNode.MatchEntrySpan(exitSpan.NextSpanId) {
			return exitSpan
		}
	}
	return nil
}

func (serviceNode *OtelServiceNode) GetEntrySpan() *OtelSpan {
	var span *OtelSpan = nil
	for _, entrySpan := range serviceNode.EntrySpans {
		if span == nil || span.Duration < entrySpan.Duration {
			span = entrySpan
		}
	}
	return span
}

func (serviceNode *OtelServiceNode) IsTopNode() bool {
	for _, entrySpan := range serviceNode.EntrySpans {
		if entrySpan.PSpanId == "" {
			return true
		}
	}
	return false
}

func (serviceNode *OtelServiceNode) GetNextServiceNode(nextSpanId string) *OtelServiceNode {
	if nextSpanId == "" || len(serviceNode.Children) == 0 {
		return nil
	}

	for _, childNode := range serviceNode.Children {
		for _, childEntry := range childNode.EntrySpans {
			if childEntry.SpanId == nextSpanId {
				return serviceNode
			}
		}
	}
	return nil
}
