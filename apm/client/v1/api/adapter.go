package api

import (
	"context"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type AdapterAPI interface {
	QueryList(ctx context.Context, params *QueryParams) ([]*model.OtelServiceNode, error)
	QueryDetail(ctx context.Context, params *QueryParams) ([]*model.OtelSpan, error)
}

type QueryParams struct {
	TraceId    string `json:"traceId"`
	ApmType    string `json:"apmType"`
	StartTime  uint64 `json:"startTime"`
	Attributes string `json:"attributes,omitempty"`
	ClusterID  string `json:"clusterId"`
}

type TraceListResponse struct {
	Success  bool                     `json:"success"`
	Data     []*model.OtelServiceNode `json:"data"`
	ErrorMsg string                   `json:"errorMsg"`
}

type TraceDetailResponse struct {
	Success  bool              `json:"success"`
	Data     []*model.OtelSpan `json:"data"`
	ErrorMsg string            `json:"errorMsg"`
}
