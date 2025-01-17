package clickhouse

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/CloudDetail/apo-module/slo/api/v1/model"
)

type SLORecordVO struct {
	EntryUri                 string    `ch:"entryUri"`
	EntryService             string    `ch:"entryService"`
	Alias                    string    `ch:"alias"`
	StartTime                int64     `ch:"startTime"`
	EndTime                  int64     `ch:"endTime"`
	RequestCount             int64     `ch:"requestCount"`
	Status                   string    `ch:"status"`
	SLOsType                 []string  `ch:"SLOs.type"`
	SLOsMultiple             []float64 `ch:"SLOs.multiple"`
	SLOsExpectedValue        []float64 `ch:"SLOs.expectedValue"`
	SLOsSource               []string  `ch:"SLOs.source"`
	SLOsCurrentValue         []float64 `ch:"SLOs.currentValue"`
	SLOsStatus               []string  `ch:"SLOs.status"`
	SlowRootCauseCountKey    []string  `ch:"slowRootCauseCount.key"`
	SlowRootCauseCountValue  []uint32  `ch:"slowRootCauseCount.value"`
	ErrorRootCauseCountKey   []string  `ch:"errorRootCauseCount.key"`
	ErrorRootCauseCountValue []uint32  `ch:"errorRootCauseCount.value"`
	Step                     string    `ch:"step"`
	IndexTimestamp           int64     `ch:"indexTimestamp"`
}

type PartFromSLOResult func(result *model.SLOResult, index int, step string, indexTimestamp int64) string

func (c *ClickhouseAPI) expandToClickhouseValue(valuesBuffer *strings.Builder, result *model.SLOResult, step string, indexTimestamp int64) int {
	for i := 0; i < len(result.SLOGroup); i++ {
		if i > 0 {
			valuesBuffer.WriteByte(',')
		}
		valuesBuffer.WriteByte('(')
		for idx, part := range c.Parts {
			if idx > 0 {
				valuesBuffer.WriteByte(',')
			}
			value := part(result, i, step, indexTimestamp)
			valuesBuffer.WriteString(value)
		}
		valuesBuffer.WriteByte(')')
	}
	return len(result.SLOGroup)
}

func getPartFromSLOResult(fieldName string) PartFromSLOResult {
	switch fieldName {
	case "entryUri":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return "'" + result.SLOServiceName.EntryUri + "'"
		}
	case "entryService":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return "'" + result.SLOServiceName.EntryService + "'"
		}
	case "alias":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return "'" + result.SLOServiceName.Alias + "'"
		}
	case "startTime":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return strconv.FormatInt(result.SLOGroup[index].StartTime, 10)
		}
	case "endTime":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return strconv.FormatInt(result.SLOGroup[index].EndTime, 10)
		}
	case "requestCount":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return strconv.Itoa(result.SLOGroup[index].RequestCount)
		}
	case "status":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			return "'" + string(result.SLOGroup[index].Status) + "'"
		}
	case "SLOs.type":
		// get Array of slos.type from SLOGroup
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			for idx, slo := range result.SLOGroup[index].SLOs {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteByte('\'')
				builder.WriteString(string(slo.SLOConfig.Type))
				builder.WriteByte('\'')
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "SLOs.multiple":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			for idx, slo := range result.SLOGroup[index].SLOs {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteString(strconv.FormatFloat(slo.SLOConfig.Multiple, 'f', 6, 64))
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "SLOs.expectedValue":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			for idx, slo := range result.SLOGroup[index].SLOs {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteString(strconv.FormatFloat(slo.SLOConfig.ExpectedValue, 'f', 2, 64))
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "SLOs.source":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			for idx, slo := range result.SLOGroup[index].SLOs {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteByte('\'')
				builder.WriteString(string(slo.SLOConfig.Source))
				builder.WriteByte('\'')
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "SLOs.currentValue":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			for idx, slo := range result.SLOGroup[index].SLOs {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteString(strconv.FormatFloat(slo.CurrentValue, 'f', 2, 64))
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "SLOs.status":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			for idx, slo := range result.SLOGroup[index].SLOs {
				if idx > 0 {
					builder.WriteByte(',')
				}
				builder.WriteByte('\'')
				builder.WriteString(string(slo.Status))
				builder.WriteByte('\'')
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "slowRootCauseCount.key":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			firstItem := true
			for key := range result.SLOGroup[index].SlowRootCauseCount {
				if !firstItem {
					builder.WriteByte(',')
				} else {
					firstItem = false
				}
				builder.WriteByte('\'')
				builder.WriteString(key)
				builder.WriteByte('\'')
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "slowRootCauseCount.value":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			firstItem := true
			for _, value := range result.SLOGroup[index].SlowRootCauseCount {
				if !firstItem {
					builder.WriteByte(',')
				} else {
					firstItem = false
				}
				builder.WriteByte('\'')
				builder.WriteString(strconv.Itoa(value))
				builder.WriteByte('\'')
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "errorRootCauseCount.key":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			firstItem := true
			for key := range result.SLOGroup[index].ErrorRootCauseCount {
				if !firstItem {
					builder.WriteByte(',')
				} else {
					firstItem = false
				}
				builder.WriteByte('\'')
				builder.WriteString(key)
				builder.WriteByte('\'')
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "errorRootCauseCount.value":
		return func(result *model.SLOResult, index int, _ string, _ int64) string {
			var builder strings.Builder
			builder.WriteByte('[')
			firstItem := true
			for _, value := range result.SLOGroup[index].ErrorRootCauseCount {
				if !firstItem {
					builder.WriteByte(',')
				} else {
					firstItem = false
				}
				builder.WriteString(strconv.Itoa(value))
			}
			builder.WriteByte(']')
			return builder.String()
		}
	case "step":
		return func(_ *model.SLOResult, _ int, step string, _ int64) string {
			return "'" + step + "'"
		}
	case "indexTimestamp":
		return func(_ *model.SLOResult, _ int, _ string, indexTimestamp int64) string {
			return strconv.FormatInt(indexTimestamp, 10)
		}
	default:
		return nil
	}
}

func buildTableFieldsDDL(t reflect.Type) (fieldNames []string, fieldsDDL string) {
	var fields []string
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Tag.Get("json")
		fieldType := getClickhouseType(field.Type)
		if fieldType == "Struct" {
			subFieldNames, subFields := buildTableFieldsDDL(field.Type)
			fields = append(fields, subFields)
			fieldNames = append(fieldNames, subFieldNames...)
		} else if fieldType == "Nested" {
			nestedFieldNames, nestedFields := buildNestedDDL(fieldName, field.Type.Elem())
			fields = append(fields, fmt.Sprintf("%s %s (%s)", fieldName, fieldType, nestedFields))
			fieldNames = append(fieldNames, nestedFieldNames...)
		} else if fieldType == "Map" {
			fields = append(fields, fmt.Sprintf("%s Nested (key String , value UInt32)", fieldName))
			fieldNames = append(fieldNames, fmt.Sprintf("%s.key", fieldName), fmt.Sprintf("%s.value", fieldName))
		} else {
			fields = append(fields, fmt.Sprintf("%s %s", fieldName, fieldType))
			fieldNames = append(fieldNames, fieldName)
		}
	}

	return fieldNames, strings.Join(fields, ", ")
}

func buildNestedDDL(parentFieldName string, t reflect.Type) (nestedFieldNames []string, fields string) {
	var nestedFields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Tag.Get("json")
		fieldType := getClickhouseType(field.Type)
		if fieldType == "Struct" {
			subFieldNames, subFields := buildTableFieldsDDL(field.Type)
			nestedFields = append(nestedFields, subFields)
			for _, subFieldName := range subFieldNames {
				nestedFieldNames = append(nestedFieldNames, fmt.Sprintf("%s.%s", parentFieldName, subFieldName))
			}
		} else if fieldType == "Nested" {
			subNestedFieldNames, subNestedFields := buildNestedDDL(fieldName, field.Type.Elem())
			nestedFields = append(nestedFields, fmt.Sprintf("%s %s (%s)", fieldName, fieldType, subNestedFields))
			nestedFieldNames = append(nestedFieldNames, subNestedFieldNames...)
		} else if fieldType == "Map" {
			nestedFields = append(nestedFields, fmt.Sprintf("%s Nested (key String , value UInt32)", fieldName))
			nestedFieldNames = append(nestedFieldNames, fmt.Sprintf("%s.key", fieldName), fmt.Sprintf("%s.value", fieldName))
		} else {
			nestedFields = append(nestedFields, fmt.Sprintf("%s %s", fieldName, fieldType))
			nestedFieldNames = append(nestedFieldNames, fmt.Sprintf("%s.%s", parentFieldName, fieldName))
		}
	}

	return nestedFieldNames, strings.Join(nestedFields, ", ")
}

func getClickhouseType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "String"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "Int64"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "UInt64"
	case reflect.Float32:
		return "Float32"
	case reflect.Float64:
		return "Float64"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Struct {
			return "Nested"
		}
		return "Array(" + getClickhouseType(t.Elem()) + ")"
	case reflect.Struct:
		return "Struct"
	case reflect.Pointer:
		return getClickhouseType(t.Elem())
	case reflect.Map:
		return "Map"
	default:
		return "String"
	}
}
