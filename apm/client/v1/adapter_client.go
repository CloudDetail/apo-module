package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CloudDetail/apo-module/apm/client/v1/api"
	"github.com/CloudDetail/apo-module/apm/model/v1"
)

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

func (c *AdapterHTTPClient) QueryList(traceId string, apmType string, startTime uint64, attributes string) ([]*model.OtelServiceNode, error) {
	queryParams := QueryParams{
		TraceId:    traceId,
		ApmType:    apmType,
		StartTime:  startTime,
		Attributes: attributes,
	}

	requestBody, err := json.Marshal(&queryParams)
	if err != nil {
		return nil, fmt.Errorf("query param is invalid, %s", err)
	}

	resp, err := c.client.Post(
		c.TraceListAddress,
		"application/json",
		bytes.NewReader(requestBody))
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
		return nil, fmt.Errorf("[x Trace NotFound] traceId: %s", traceId)
	}
	return response.Data, nil
}

func (c *AdapterHTTPClient) QueryDetail(traceId string, apmType string, startTime uint64, attributes string) ([]*model.OtelSpan, error) {
	queryParams := QueryParams{
		TraceId:    traceId,
		ApmType:    apmType,
		StartTime:  startTime,
		Attributes: attributes,
	}

	requestBody, err := json.Marshal(&queryParams)
	if err != nil {
		return nil, fmt.Errorf("query param is invalid, %s", err)
	}

	resp, err := c.client.Post(
		c.TraceDetailAddress,
		"application/json",
		bytes.NewReader(requestBody))
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
