package elasticsearch

import (
	"strings"
	"time"
)

const SLORecordIndex = "slo_result"
const DefaultTimePattern = "20060102"

var SLORecordESPattern = NewElasticSearchIndexPattern("", SLORecordIndex, "", DefaultTimePattern)

type ElasticSearchIndexPattern struct {
	Prefix      string
	Index       string
	Suffix      string
	TimePattern string

	base string
}

func NewElasticSearchIndexPattern(prefix string, index string, suffix string, timePattern string) *ElasticSearchIndexPattern {
	pattern := &ElasticSearchIndexPattern{
		Prefix:      prefix,
		Index:       index,
		Suffix:      suffix,
		TimePattern: timePattern,
	}
	pattern.buildBase()
	return pattern
}

func (p *ElasticSearchIndexPattern) GetIndexWithTimePattern(timestampMillis int64) (string, error) {
	var pattern string
	if len(p.TimePattern) > 0 {
		pattern = time.UnixMilli(int64(timestampMillis)).Format(p.TimePattern)
	}

	var builder strings.Builder
	builder.WriteString(p.base)
	builder.WriteByte('-')
	builder.WriteString(pattern)
	return builder.String(), nil
}

func (p *ElasticSearchIndexPattern) GetSearchIndexPattern() (string, error) {
	var builder strings.Builder
	builder.WriteString(p.base)
	builder.WriteString("-*")
	return builder.String(), nil
}

func (p *ElasticSearchIndexPattern) buildBase() error {
	if len(p.Index) == 0 {
		return &ErrInvalidIndex{Index: p.Index}
	}
	var builder strings.Builder
	if len(p.Prefix) > 0 {
		builder.WriteString(p.Prefix)
		builder.WriteByte('-')
	}
	builder.WriteString(p.Index)
	if len(p.Suffix) > 0 {
		builder.WriteByte('-')
		builder.WriteString(p.Suffix)
	}
	p.base = builder.String()
	return nil
}
