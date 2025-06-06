package api

import (
	"context"

	"github.com/CloudDetail/apo-module/apm/model/v1"
)

type AdapterAPI interface {
	QueryList(traceId string, apmType string, startTime uint64, attributes string) ([]*model.OtelServiceNode, error)
	QueryDetail(traceId string, apmType string, startTime uint64, attributes string) ([]*model.OtelSpan, error)

	QueryListWithCtx(ctx context.Context, traceId string, apmType string, startTime uint64, attributes string) ([]*model.OtelServiceNode, error)
	QueryDetailWithCtx(ctx context.Context, traceId string, apmType string, startTime uint64, attributes string) ([]*model.OtelSpan, error)
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
