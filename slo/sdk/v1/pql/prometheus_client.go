package pql

import (
	"context"
	"log"
	"math"
	"time"

	promApi "github.com/prometheus/client_golang/api/prometheus/v1"
	promModel "github.com/prometheus/common/model"
)

type PrometheusClient struct {
	promApi.API
}

func NewPrometheusClient(api promApi.API) *PrometheusClient {
	return &PrometheusClient{API: api}
}

func (client *PrometheusClient) BucketLabelName() string {
	return "le"
}

func (client *PrometheusClient) QueryMetric(endTime uint64, query string) (float64, error) {
	if query == "" {
		return 0, nil
	}
	result, warnings, err := client.Query(context.Background(), query, time.UnixMilli(int64(endTime/1e6)))
	if err != nil {
		return 0, err
	}
	if len(warnings) > 0 {
		log.Printf("Request Prometheus Warning: %s", warnings)
	}
	if vector, ok := result.(promModel.Vector); ok {
		if len(vector) != 1 {
			log.Printf("[x Query Metrics] %s, Size: %d", query, len(vector))
		} else {
			if math.IsNaN(float64(vector[0].Value)) {
				return 0, nil
			}
			return float64(vector[0].Value), nil
		}
	}
	return 0, nil
}

func (client *PrometheusClient) QueryVectorMillTS(endTimeMill int64, query string) (promModel.Vector, error) {
	if query == "" {
		return nil, nil
	}
	result, warnings, err := client.Query(context.Background(), query, time.UnixMilli(endTimeMill))
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Printf("Request Prometheus Warning: %s", warnings)
	}
	return result.(promModel.Vector), err
}

func (client *PrometheusClient) QueryMetricMillTS(endTime int64, query string) (float64, error) {
	if query == "" {
		return 0, nil
	}
	result, warnings, err := client.Query(context.Background(), query, time.UnixMilli(endTime))
	if err != nil {
		return 0, err
	}
	if len(warnings) > 0 {
		log.Printf("Request Prometheus Warning: %s", warnings)
	}
	if vector, ok := result.(promModel.Vector); ok {
		if len(vector) != 1 {
			log.Printf("[x Query Metrics] %s, Size: %d", query, len(vector))
		} else {
			if math.IsNaN(float64(vector[0].Value)) {
				return 0, nil
			}
			return float64(vector[0].Value), nil
		}
	}
	return 0, nil
}
func (client *PrometheusClient) QueryTimeSeriesMatrix(timeRange promApi.Range, query string) (promModel.Matrix, error) {
	if query == "" {
		return nil, nil
	}

	result, warnings, err := client.QueryRange(context.Background(), query, timeRange)
	if err != nil {
		return nil, err
	}
	if len(warnings) > 0 {
		log.Printf("Request Prometheus Warning: %s", warnings)
	}
	return result.(promModel.Matrix), nil
}
