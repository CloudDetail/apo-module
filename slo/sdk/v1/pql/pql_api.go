package pql

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/api"
	promV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PQLApi interface {
	QueryTimeSeriesMatrix(timeRange promV1.Range, query string) (model.Matrix, error)
	QueryVectorMillTS(endTimeMills int64, query string) (model.Vector, error)
	QueryMetric(endTime uint64, query string) (float64, error)
	QueryMetricMillTS(endTimeMills int64, query string) (float64, error)
	BucketLabelName() string
}

func NewPQLApi(address string, pqlType string) (PQLApi, error) {
	client, err := api.NewClient(api.Config{
		Address: address,
	})
	if err != nil {
		return nil, err
	}
	api := promV1.NewAPI(client)
	switch pqlType {
	case "vm":
		log.Printf("[SLO setup] query from Victoria Metrics: addr: %s", address)
		return &VictoriaMetricsClient{
			PrometheusClient: &PrometheusClient{API: api},
		}, nil
	case "prom":
		log.Printf("[SLO setup] query from Prometheus: addr: %s", address)
		return &PrometheusClient{API: api}, nil
	default:
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		buildInfo, err := api.Buildinfo(ctx)
		if err != nil {
			log.Printf("[SLO setup] can not check PQL API Type, use Victoria API as default: prometheus Server is not ready yet: %v ", err)
			break // break switch
		}

		resp, err := http.Get(fmt.Sprintf("%s/vmui/?", address))
		if err != nil {
			break // break switch
		}
		if resp.StatusCode != 200 {
			// Can not check whether is victoria metrics, use prometheus as default
			log.Printf("[SLO setup] query from Prometheus: addr: %s, version: %s", address, buildInfo.Version)
			return &PrometheusClient{API: api}, nil
		}
	}

	log.Printf("[SLO setup] query from Victoria Metrics: addr: %s", address)
	return &VictoriaMetricsClient{
		PrometheusClient: &PrometheusClient{API: api},
	}, nil
}
