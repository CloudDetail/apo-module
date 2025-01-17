package checker

import (
	"fmt"
	"time"

	"github.com/CloudDetail/apo-module/slo/api/v1"
	"github.com/CloudDetail/apo-module/slo/api/v1/model"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/pql"
	prom "github.com/prometheus/client_golang/api/prometheus/v1"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	pmodel "github.com/prometheus/common/model"
	"go.uber.org/multierr"
)

var _ api.Checker = &PrometheusChecker{}

var DefaultChecker api.Checker

type PrometheusChecker struct {
	pql.PQLApi
}

func NewPrometheusChecker(pqlApi pql.PQLApi) *PrometheusChecker {
	return &PrometheusChecker{PQLApi: pqlApi}
}

func (p *PrometheusChecker) GetTimeSeriesGroupResult(key *model.SLOEntryKey, sloConfigs []model.SLOConfig, startTime int64, endTime int64, step time.Duration) ([]model.SLOGroup, error) {
	if key == nil {
		return nil, &model.ErrInvalidSLOKey{SloKey: key}
	}

	searchStart := startTime + int64(step)/1e6
	timeRange := prom.Range{
		Start: time.UnixMilli(searchStart),
		End:   time.UnixMilli(endTime),
		Step:  step,
	}

	slots := NewSLOTimeSeries(startTime, endTime, step)

	durationNanos := (endTime - startTime) * 1e6
	if durationNanos < int64(time.Minute) {
		return slots.Res, fmt.Errorf("search range (%d,%d) is too small, must large than 1minute", startTime, endTime)
	}

	var err error
	var totalCount int64 = 0
	duration := pql.GetDurationFromStep(step)
	matrix, err := p.QueryTimeSeriesMatrix(timeRange, pql.GetRequestCountIncreasedPQL(key.EntryURI, duration))
	if err == nil && len(matrix) > 0 {
		for _, tsValue := range matrix[0].Values {
			totalCount += int64(tsValue.Value)
			slots.AddRequestCountAtTimestamp(float64(tsValue.Value), int64(tsValue.Timestamp))
		}
	}

	if totalCount == 0 {
		return slots.Res, model.ErrNotActiveUriError
	}

	var errorsDuringSearch error
	for i := 0; i < len(sloConfigs); i++ {
		sloConfig := sloConfigs[i]
		var matrix pmodel.Matrix
		var err error
		switch sloConfig.Type {
		case model.SLO_LATENCY_P90_TYPE:
			matrix, err = p.QueryTimeSeriesMatrix(timeRange, pql.GetLatencyPercentilePQL(0.90, key.EntryURI, duration, p.BucketLabelName()))
		case model.SLO_LATENCY_P95_TYPE:
			matrix, err = p.QueryTimeSeriesMatrix(timeRange, pql.GetLatencyPercentilePQL(0.95, key.EntryURI, duration, p.BucketLabelName()))
		case model.SLO_LATENCY_P99_TYPE:
			matrix, err = p.QueryTimeSeriesMatrix(timeRange, pql.GetLatencyPercentilePQL(0.99, key.EntryURI, duration, p.BucketLabelName()))
		case model.SLO_SUCCESS_RATE_TYPE:
			matrix, err = p.QueryTimeSeriesMatrix(timeRange, pql.GetSuccessRatePQL(key.EntryURI, duration))
		default:
			errorsDuringSearch = multierr.Append(errorsDuringSearch, &model.ErrInvalidSLOType{SloType: sloConfig.Type})
			continue
		}

		if err != nil {
			errorsDuringSearch = multierr.Append(errorsDuringSearch, fmt.Errorf("failed to query %s for Entry: %v, err: %w", sloConfig.Type, key, err))
			slots.FinishUpdateSLO()
			continue
		}

		if len(matrix) < 1 && sloConfig.Type == model.SLO_SUCCESS_RATE_TYPE {
			slots.AllRequestCountFailed(&sloConfig)
			continue
		}

		if len(matrix) < 1 {
			errorsDuringSearch = multierr.Append(errorsDuringSearch, fmt.Errorf("null of %s get for Entry: %v", sloConfig.Type, key))
			slots.FinishUpdateSLO()
			continue
		}

		for _, tsValue := range matrix[0].Values {
			slo := checkResult(sloConfig, float64(tsValue.Value))
			slots.AddSLOAtTimestamp(slo, int64(tsValue.Timestamp))
		}

		slots.FinishUpdateSLO()
	}
	return slots.Res, errorsDuringSearch
}

func (p *PrometheusChecker) GetHistorySLO(key model.SLOEntryKey, endTime int64) (model.SLOHistory, error) {
	sloHistory := model.SLOHistory{
		SuccessRate: map[string]float64{},
		Latency:     map[string]model.HistoryLatency{},
	}

	if successRates, err := p.getSuccessRates(key, endTime); err != nil {
		return sloHistory, err
	} else {
		sloHistory.SuccessRate = successRates
	}

	todayTsMill := getDayUnixMilli(endTime)
	latencies := map[string]model.HistoryLatency{
		"P90": p.getP9xs(key, todayTsMill, endTime, 0.9, 500.0),
		"P95": p.getP9xs(key, todayTsMill, endTime, 0.95, 500.0),
		"P99": p.getP9xs(key, todayTsMill, endTime, 0.99, 500.0),
	}
	sloHistory.Latency = latencies
	return sloHistory, nil
}

func (p *PrometheusChecker) ListEntryTemp(service string, startTimeMill int64, endTimeMill int64) (targets []model.SLOEntryKeyTemp, err error) {
	duration := pql.GetDurationFromNS((endTimeMill - startTimeMill) * 1e6)
	if duration == "1m" {
		duration = "5m"
	}
	targets = make([]model.SLOEntryKeyTemp, 0)

	var filters []string = make([]string, 0)
	if len(service) > 0 {
		filters = append(filters, fmt.Sprintf(`content_key=~".*%s.*"`, service))
	}
	timeSeries, err := p.QueryTimeSeriesMatrix(v1.Range{
		Start: time.UnixMilli(startTimeMill),
		End:   time.UnixMilli(endTimeMill),
		Step:  time.Second * 15,
	}, pql.GetEntryGroupTemp(duration, filters...))

	for _, matrix := range timeSeries {
		contentKey, find := matrix.Metric["content_key"]
		if !find {
			continue
		}
		serviceName, find := matrix.Metric["svc_name"]
		if !find {
			continue
		}
		targets = append(targets, model.SLOEntryKeyTemp{
			EntryURI:     string(contentKey),
			EntryService: string(serviceName),
		})
	}
	return targets, err
}

func (p *PrometheusChecker) ListContentKeyTemp(service string, startTimeMill int64, endTimeMill int64) (targets []model.SLOEntryKeyTemp, err error) {
	duration := pql.GetDurationFromNS((endTimeMill - startTimeMill) * 1e6)
	if duration == "1m" {
		duration = "5m"
	}
	targets = make([]model.SLOEntryKeyTemp, 0)

	var filters []string = make([]string, 0)
	if len(service) > 0 {
		filters = append(filters, fmt.Sprintf(`content_key=~".*%s.*"`, service))
	}
	timeSeries, err := p.QueryTimeSeriesMatrix(v1.Range{
		Start: time.UnixMilli(startTimeMill),
		End:   time.UnixMilli(endTimeMill),
		Step:  time.Second * 15,
	}, pql.GetContentKeyGroupTemp(duration, filters...))

	for _, matrix := range timeSeries {
		contentKey, find := matrix.Metric["content_key"]
		if !find {
			continue
		}
		serviceName, find := matrix.Metric["svc_name"]
		if !find {
			continue
		}
		targets = append(targets, model.SLOEntryKeyTemp{
			EntryURI:     string(contentKey),
			EntryService: string(serviceName),
		})
	}
	return targets, err
}

func (p *PrometheusChecker) ListEntry(service string, startTimeMill int64, endTimeMill int64) (targets []model.SLOEntryKey, err error) {
	duration := pql.GetDurationFromNS((endTimeMill - startTimeMill) * 1e6)
	if duration == "1m" {
		duration = "5m"
	}
	targets = make([]model.SLOEntryKey, 0)

	var filters []string = make([]string, 0)
	if len(service) > 0 {
		filters = append(filters, fmt.Sprintf(`content_key=~".*%s.*"`, service))
	}

	timeSeries, err := p.QueryTimeSeriesMatrix(v1.Range{
		Start: time.UnixMilli(startTimeMill),
		End:   time.UnixMilli(endTimeMill),
		Step:  time.Second * 15,
	}, pql.GetEntryGroup(duration, filters...))

	for _, matrix := range timeSeries {
		contentKey, find := matrix.Metric["content_key"]
		if !find {
			continue
		}
		// serviceName, find := group.Metric["svc_name"]
		// if !find {
		// 	continue
		// }
		targets = append(targets, model.SLOEntryKey{
			EntryURI: string(contentKey),
		})
	}
	return targets, err
}

const (
	dayDuration  = "24h"
	hourDuration = "1h"
)

func (p *PrometheusChecker) getP9xs(key model.SLOEntryKey, todayTsMill int64, nowTsMill int64, percentile float64, defaultValue float64) model.HistoryLatency {
	latency := model.HistoryLatency{}
	if value, err := p.QueryMetricMillTS(todayTsMill, pql.GetLatencyPercentilePQL(percentile, key.EntryURI, dayDuration, p.BucketLabelName())); err == nil && value != 0 {
		latency.Range = "yesterday"
		latency.Value = value / 1e6
		return latency
	}

	if value, err := p.QueryMetricMillTS(nowTsMill, pql.GetLatencyPercentilePQL(percentile, key.EntryURI, hourDuration, p.BucketLabelName())); err == nil && value != 0 {
		latency.Range = "last1h"
		latency.Value = value / 1e6
		return latency
	}

	latency.Range = "constant"
	latency.Value = defaultValue
	return latency
}

func (p *PrometheusChecker) getSuccessRates(key model.SLOEntryKey, endTime int64) (map[string]float64, error) {
	values := make(map[string]float64)

	if successRate1h, err := p.QueryMetricMillTS(endTime, pql.GetSuccessRatePQL(key.EntryURI, "1h")); err != nil {
		return values, err
	} else {
		values["last1h"] = successRate1h
	}

	if successRate12h, err := p.QueryMetricMillTS(endTime, pql.GetSuccessRatePQL(key.EntryURI, "12h")); err != nil {
		return values, err
	} else {
		values["last12h"] = successRate12h
	}

	if successRate24h, err := p.QueryMetricMillTS(endTime, pql.GetSuccessRatePQL(key.EntryURI, "24h")); err != nil {
		return values, err
	} else {
		values["last24h"] = successRate24h
	}

	if successRate48h, err := p.QueryMetricMillTS(endTime, pql.GetSuccessRatePQL(key.EntryURI, "48h")); err != nil {
		return values, err
	} else {
		values["last48h"] = successRate48h
	}

	return values, nil
}

func GetHistorySLO(key model.SLOEntryKey, endTime int64) (model.SLOHistory, error) {
	return DefaultChecker.GetHistorySLO(key, endTime)
}
