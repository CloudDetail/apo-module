package client

import (
	"context"
	"errors"
	"fmt"

	"github.com/CloudDetail/apo-module/apm/client/v1/api"
	"github.com/CloudDetail/apo-module/model/v1"

	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
)

var _ api.ApmTraceAPI = &ApmTraceClient{}

var (
	ErrUnknownApmType error = errors.New("no match apmType is found")
)

type ApmTraceClient struct {
	api            api.AdapterAPI
	muatedRatio    int
	mutateNodeMode string
	getDetailTypes []string
}

func NewApmTraceClient(address string, timeout int64, muatedRatio int, mutateNodeMode string, getDetailTypes []string) *ApmTraceClient {
	return &ApmTraceClient{
		api:            NewAdapterHTTPClient(address, timeout),
		muatedRatio:    muatedRatio,
		mutateNodeMode: mutateNodeMode,
		getDetailTypes: getDetailTypes,
	}
}

func NewApmTraceClientByAPI(api api.AdapterAPI, muatedRatio int, mutateNodeMode string, getDetailTypes []string) *ApmTraceClient {
	return &ApmTraceClient{
		api:            api,
		muatedRatio:    muatedRatio,
		mutateNodeMode: mutateNodeMode,
		getDetailTypes: getDetailTypes,
	}
}

func (client *ApmTraceClient) QueryServices(ctx context.Context, clusterID string, apmType string, traceId string, rootTrace *model.TraceLabels) ([]*apmmodel.OtelServiceNode, error) {
	param := &api.QueryParams{
		TraceId:    traceId,
		ApmType:    apmType,
		StartTime:  rootTrace.StartTime / 1e6,
		Attributes: rootTrace.Attributes,
		ClusterID:  clusterID,
	}
	return client.api.QueryList(ctx, param)
}

func (client *ApmTraceClient) QueryTrace(ctx context.Context, clusterID string, apmType string, traceId string, rootTrace *model.TraceLabels) (*apmmodel.OTelTrace, error) {
	param := &api.QueryParams{
		TraceId:    traceId,
		ApmType:    apmType,
		StartTime:  rootTrace.StartTime / 1e6,
		Attributes: rootTrace.Attributes,
		ClusterID:  clusterID,
	}

	serviceNodes, err := client.api.QueryList(ctx, param)
	if err != nil {
		return nil, err
	}
	apmTrace := apmmodel.NewOTelTrace(apmType)
	for _, serviceNode := range serviceNodes {
		apmTrace.AddServiceNode(serviceNode, nil)
	}

	if err := apmTrace.CheckRoot(serviceNodes); err != nil {
		return nil, err
	}

	rootService := apmTrace.GetRoot()
	if rootService == nil {
		if apmType != "arms" {
			return nil, fmt.Errorf("miss RootSpan")
		}
		code := apmmodel.StatusCodeOk
		if rootTrace.IsError {
			code = apmmodel.StatusCodeError
		}
		rootSpan := &apmmodel.OtelSpan{
			StartTime:   rootTrace.StartTime, // ns
			Duration:    rootTrace.Duration,  // ns
			ServiceName: rootTrace.ServiceName,
			Name:        rootTrace.Url,
			SpanId:      rootTrace.ApmSpanId,
			PSpanId:     "",
			Kind:        apmmodel.SpanKindServer,
			Code:        code,
			NotSampled:  !rootTrace.IsSampled,
			Attributes:  make(map[string]string, 0),
		}
		rootService = &apmmodel.OtelServiceNode{
			StartTime:  rootSpan.StartTime,
			EntrySpans: []*apmmodel.OtelSpan{rootSpan},
			VNode:      false,
		}
		apmTrace.AddServiceNode(rootService, nil)
		apmTrace.SetRoot(rootService)

		for _, serviceNode := range serviceNodes {
			if serviceNode.Parent == nil {
				rootService.Children = append(rootService.Children, serviceNode)
				serviceNode.CheckVNode(rootService.SpanId)
			}
		}
	}
	rootService.SetFixTime()
	return apmTrace, nil
}

func (client *ApmTraceClient) FillMutatedSpan(ctx context.Context, clusterID string, apmType string, traceId string, serviceNode *apmmodel.OtelServiceNode) error {
	param := &api.QueryParams{
		TraceId:    traceId,
		ApmType:    apmType,
		StartTime:  serviceNode.GetStartTime() / 1e6,
		Attributes: serviceNode.Attribute,
		ClusterID:  clusterID,
	}

	spans, err := client.api.QueryDetail(ctx, param)
	if err != nil {
		return err
	}
	for _, span := range spans {
		if span.IsError() {
			serviceNode.ErrorSpans = append(serviceNode.ErrorSpans, span)
		}
		if span.Kind.IsExit() {
			serviceNode.ExitSpans = append(serviceNode.ExitSpans, span)
		}
	}
	return nil
}

func (client *ApmTraceClient) QueryMutatedSlowTraceTree(ctx context.Context, clusterID string, traceId string, traces *model.Traces) (*model.TraceTreeNode, []*model.ApmClientCall, error) {
	entryTrace := traces.RootTrace.Labels
	if uint64(entryTrace.ThresholdValue) >= entryTrace.Duration {
		return nil, nil, fmt.Errorf("entry service(%s) duration(%d) is less than threshold(%s(%s)=%f)",
			entryTrace.ServiceName, entryTrace.Duration, entryTrace.ThresholdType, entryTrace.ThresholdRange,
			entryTrace.ThresholdValue)
	}

	apmType := entryTrace.ApmType
	apmTrace, err := client.QueryTrace(ctx, clusterID, apmType, traceId, entryTrace)
	if err != nil {
		return nil, nil, err
	}

	apmTraceTree, err := BuildTopologyTree(apmTrace, traces)
	if err != nil {
		return nil, nil, err
	}

	mutatedTrace, err := apmTraceTree.GetMutatedTraceNode(traceId, client.muatedRatio, client.mutateNodeMode)
	if err != nil {
		return nil, nil, err
	}

	if client.NeedGetDetailSpan(ctx, apmType) {
		if err := client.FillMutatedSpan(ctx, clusterID, apmType, traceId, apmTrace.GetServiceNode(mutatedTrace.SpanId)); err != nil {
			return nil, nil, err
		}
	}

	clientCalls := GetClientCalls(apmTrace, mutatedTrace.SpanId)
	return apmTraceTree.Root, clientCalls, nil
}

func (client *ApmTraceClient) QueryErrorTraceTree(ctx context.Context, clusterID string, traceId string, traces *model.Traces) (*model.ErrorTreeNode, error) {
	entryTrace := traces.RootTrace.Labels
	apmTrace, err := client.QueryTrace(ctx, clusterID, entryTrace.ApmType, traceId, entryTrace)
	if err != nil {
		return nil, err
	}

	apmErrorTree, err := BuildErrorTree(apmTrace, traces)
	if err != nil {
		return nil, err
	}
	if client.NeedGetDetailSpan(ctx, entryTrace.ApmType) {
		for spanId, errorNode := range apmErrorTree.NodeMap {
			if errorNode.IsError && errorNode.IsSampled {
				if node := apmTrace.GetServiceNode(spanId); node != nil {
					if err := client.FillMutatedSpan(ctx, clusterID, entryTrace.ApmType, traces.TraceId, node); err != nil {
						return nil, err
					}
					errorNode.ErrorSpans = GetErrorSpans(node)
				}
			}
		}
	}
	if _, err := apmErrorTree.GetRootCauseErrorNode(traceId); err != nil {
		return nil, err
	}

	return apmErrorTree.Root, nil
}

func (client *ApmTraceClient) NeedGetDetailSpan(ctx context.Context, apmType string) bool {
	if len(client.getDetailTypes) == 0 {
		return false
	}
	for _, detailType := range client.getDetailTypes {
		if detailType == apmType {
			return true
		}
	}
	return false
}
