package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudDetail/apo-module/apm/client/v1/api"
	"github.com/CloudDetail/apo-module/apm/model/v1"
)

var _ api.AdapterAPI = &AdapterHTTPClient{}

type AdapterHTTPClient struct {
	TraceListAddress   string
	TraceDetailAddress string
	Timeout            time.Duration

	client *http.Client
}

type QueryParams struct {
	TraceId    string `json:"traceId"`
	ApmType    string `json:"apmType"`
	StartTime  uint64 `json:"startTime"`
	Attributes string `json:"attributes,omitempty"`

	ClusterID string `json:"clusterId"`
}

func NewAdapterHTTPClient(address string, timeout int64) *AdapterHTTPClient {
	timeoutDuration := time.Duration(timeout) * time.Second
	return &AdapterHTTPClient{
		TraceListAddress:   fmt.Sprintf("http://%s/trace/list", address),
		TraceDetailAddress: fmt.Sprintf("http://%s/trace/detail", address),
		Timeout:            timeoutDuration,
		client:             &http.Client{Timeout: timeoutDuration},
	}
}

func (c *AdapterHTTPClient) SetRountTriper(rt http.RoundTripper) {
	c.client.Transport = rt
}

func (c *AdapterHTTPClient) QueryList(ctx context.Context, queryParams *api.QueryParams) ([]*model.OtelServiceNode, error) {
	requestBody, err := json.Marshal(&queryParams)
	if err != nil {
		return nil, fmt.Errorf("query param is invalid, %s", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TraceListAddress, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.TraceListResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New(response.ErrorMsg)
	}
	if len(response.Data) == 0 {
		return nil, fmt.Errorf("[x Trace NotFound] traceId: %s", queryParams.TraceId)
	}
	return response.Data, nil
}

func (c *AdapterHTTPClient) QueryDetail(ctx context.Context, queryParams *api.QueryParams) ([]*model.OtelSpan, error) {
	requestBody, err := json.Marshal(&queryParams)
	if err != nil {
		return nil, fmt.Errorf("query param is invalid, %s", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.TraceDetailAddress, bytes.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.TraceDetailResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}
	if !response.Success {
		return nil, errors.New(response.ErrorMsg)
	}
	return response.Data, nil
}
