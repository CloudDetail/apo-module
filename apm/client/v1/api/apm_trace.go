package api

import (
	"context"

	apmmodel "github.com/CloudDetail/apo-module/apm/model/v1"
	"github.com/CloudDetail/apo-module/model/v1"
)

type ApmTraceAPI interface {
	QueryServices(ctx context.Context, clusterID string, apmType string, traceId string, rootTrace *model.TraceLabels) ([]*apmmodel.OtelServiceNode, error)
	QueryTrace(ctx context.Context, clusterID string, apmType string, traceId string, rootTrace *model.TraceLabels) (*apmmodel.OTelTrace, error)
	FillMutatedSpan(ctx context.Context, clusterID string, apmType string, traceId string, serviceNode *apmmodel.OtelServiceNode) error
	QueryMutatedSlowTraceTree(ctx context.Context, clusterID string, traceId string, traces *model.Traces) (*model.TraceTreeNode, []*model.ApmClientCall, error)
	QueryErrorTraceTree(ctx context.Context, clusterID string, traceId string, traces *model.Traces) (*model.ErrorTreeNode, error)
	NeedGetDetailSpan(ctx context.Context, apmType string) bool
}
