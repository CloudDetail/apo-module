package api

import (
	"context"

	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/CloudDetail/apo-module/model/v1"
)

type ApmTraceAPI interface {
	QueryServices(apmType string, traceId string, rootTrace *model.TraceLabels) ([]*apmmodel.OtelServiceNode, error)
	QueryTrace(apmType string, traceId string, rootTrace *model.TraceLabels) (*apmmodel.OTelTrace, error)
	FillMutatedSpan(apmType string, traceId string, serviceNode *apmmodel.OtelServiceNode) error
	QueryMutatedSlowTraceTree(traceId string, traces *model.Traces) (*model.TraceTreeNode, []*model.ApmClientCall, error)
	QueryErrorTraceTree(traceId string, traces *model.Traces) (*model.ErrorTreeNode, error)
	NeedGetDetailSpan(apmType string) bool

	QueryServicesWithCtx(ctx context.Context, apmType string, traceId string, rootTrace *model.TraceLabels) ([]*apmmodel.OtelServiceNode, error)
	QueryTraceWithCtx(ctx context.Context, apmType string, traceId string, rootTrace *model.TraceLabels) (*apmmodel.OTelTrace, error)
	FillMutatedSpanWithCtx(ctx context.Context, apmType string, traceId string, serviceNode *apmmodel.OtelServiceNode) error
	QueryMutatedSlowTraceTreeWithCtx(ctx context.Context, traceId string, traces *model.Traces) (*model.TraceTreeNode, []*model.ApmClientCall, error)
	QueryErrorTraceTreeWithCtx(ctx context.Context, traceId string, traces *model.Traces) (*model.ErrorTreeNode, error)
	NeedGetDetailSpanWithCtx(ctx context.Context, apmType string) bool
}
