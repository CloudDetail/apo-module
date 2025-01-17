package pql

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	LatencyPercentilePQLTemplate    = `histogram_quantile(%f,sum(increase(kindling_span_trace_duration_nanoseconds_bucket{content_key="%s"}[%s])) by (%s))`
	SuccessRatePQLTemplate          = `sum (increase(kindling_span_trace_duration_nanoseconds_count{content_key="%s",is_error="false"}[%s]))/sum (increase(kindling_span_trace_duration_nanoseconds_count{content_key="%s"}[%s])) * 100`
	RequestCountIncreasePQLTemplate = `sum(increase(kindling_span_trace_duration_nanoseconds_count{content_key="%s"}[%s]))`

	EntryGroupPQLTemplate           = `increase (sum by (content_key) (kindling_span_trace_duration_nanoseconds_count{top_span='true'})[%s]) >0 `
	EntryGroupPQLTemplateWithFilter = `increase (sum by (content_key) (kindling_span_trace_duration_nanoseconds_count{top_span='true',%s})[%s]) >0 `

	EntryGroupPQLTemplateTemp           = `increase (sum by (content_key,svc_name) (kindling_span_trace_duration_nanoseconds_count{top_span='true'})[%s]) >0 `
	EntryGroupPQLTemplateWithFilterTemp = `increase (sum by (content_key,svc_name) (kindling_span_trace_duration_nanoseconds_count{top_span='true',%s})[%s]) >0 `

	ContentKeyGroupPQLTemplateTemp           = `increase (sum by (content_key,svc_name) (kindling_span_trace_duration_nanoseconds_count)[%s]) >0 `
	ContentKeyGroupPQLTemplateWithFilterTemp = `increase (sum by (content_key,svc_name) (kindling_span_trace_duration_nanoseconds_count{%s})[%s]) >0 `
)

func GetEntryGroup(duration string, filters ...string) string {
	if len(filters) < 1 {
		return fmt.Sprintf(EntryGroupPQLTemplate, duration)
	}
	return fmt.Sprintf(EntryGroupPQLTemplateWithFilter, strings.Join(filters, ","), duration)
}

func GetEntryGroupTemp(duration string, filters ...string) string {
	if len(filters) < 1 {
		return fmt.Sprintf(EntryGroupPQLTemplateTemp, duration)
	}
	return fmt.Sprintf(EntryGroupPQLTemplateWithFilterTemp, strings.Join(filters, ","), duration)
}

func GetContentKeyGroupTemp(duration string, filters ...string) string {
	if len(filters) < 1 {
		return fmt.Sprintf(ContentKeyGroupPQLTemplateTemp, duration)
	}
	return fmt.Sprintf(ContentKeyGroupPQLTemplateWithFilterTemp, strings.Join(filters, ","), duration)
}

func GetLatencyPercentilePQL(percentile float64, content_key string, duration string, bucketLabelName string) string {
	return fmt.Sprintf(LatencyPercentilePQLTemplate, percentile, content_key, duration, bucketLabelName)
}

func GetRequestCountIncreasedPQL(content_key string, duration string) string {
	return fmt.Sprintf(RequestCountIncreasePQLTemplate, content_key, duration)
}

func GetSuccessRatePQL(content_key string, duration string) string {
	//duration := "1m"
	return fmt.Sprintf(SuccessRatePQLTemplate, content_key, duration, content_key, duration)
}

func GetDurationFromStep(step time.Duration) string {
	var stepNS = step.Nanoseconds()
	if stepNS > int64(time.Hour) {
		return strconv.FormatInt(stepNS/int64(time.Hour), 10) + "h"
	}

	if stepNS > int64(time.Minute) {
		return strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
	}

	return "1m"
}

func GetDurationFromNS(stepNS int64) string {
	if stepNS > int64(time.Hour) {
		return strconv.FormatInt(stepNS/int64(time.Hour), 10) + "h"
	}

	if stepNS > int64(time.Minute) {
		return strconv.FormatInt(stepNS/int64(time.Minute), 10) + "m"
	}

	return "1m"
}
