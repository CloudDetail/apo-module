package model

import (
	"sync"
	"time"
)

type OnOffMetrics struct {
	TraceId    string
	Metrics    map[string]string // <spanId, metrics>
	ExpireTime int64
	mutex      sync.RWMutex
}

func NewOnOffMetrics(metric *OnOffMetricGroup, cacheTime int64) *OnOffMetrics {
	return &OnOffMetrics{
		TraceId: metric.TraceId,
		Metrics: map[string]string{
			metric.SpanId: metric.Metrics,
		},
		ExpireTime: time.Now().Unix() + cacheTime,
	}
}

func (m *OnOffMetrics) AddMetric(spanId string, metrics string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.Metrics[spanId] = metrics
}

func (m *OnOffMetrics) RemoveMetric(spanId string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if metrics, ok := m.Metrics[spanId]; ok {
		delete(m.Metrics, spanId)
		return metrics
	}
	return ""
}

type OnOffMetricGroup struct {
	TraceId string `json:"trace_id"`
	SpanId  string `json:"span_id"`
	Metrics string `json:"metrics"`
}
