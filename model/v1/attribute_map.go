package model

import (
	"encoding/json"
	"strconv"
)

type AttributeValueType int

const (
	StringAttributeValueType AttributeValueType = iota
	IntAttributeValueType
	BooleanAttributeValueType
)

type AttributeMap struct {
	values map[string]AttributeValue
}

func (a AttributeMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.values)
}

func NewAttributeMap() *AttributeMap {
	values := make(map[string]AttributeValue)
	return &AttributeMap{values}
}

func (a *AttributeMap) Size() int {
	return len(a.values)
}

func (a *AttributeMap) HasAttribute(key string) bool {
	_, existing := a.values[key]
	return existing
}

func (a *AttributeMap) GetStringValue(key string) string {
	value := a.values[key]
	if x, ok := value.(*stringValue); ok {
		return x.value
	}
	return ""
}

func (a *AttributeMap) AddStringValue(key string, value string) {
	a.values[key] = &stringValue{
		value: value,
	}
}

func (a *AttributeMap) GetIntValue(key string) int64 {
	value := a.values[key]
	if x, ok := value.(*intValue); ok {
		return x.value
	}
	return 0
}

func (a *AttributeMap) AddIntValue(key string, value int64) {
	a.values[key] = &intValue{
		value: value,
	}
}

func (a *AttributeMap) GetBoolValue(key string) bool {
	value := a.values[key]
	if x, ok := value.(*boolValue); ok {
		return x.value
	}
	return false
}

func (a *AttributeMap) AddBoolValue(key string, value bool) {
	a.values[key] = &boolValue{
		value: value,
	}
}

func (a *AttributeMap) ToStringMap() map[string]string {
	stringMap := make(map[string]string)
	if a == nil {
		return stringMap
	}
	for k, v := range a.values {
		stringMap[k] = v.ToString()
	}
	return stringMap
}

func (a *AttributeMap) String() string {
	json, _ := json.Marshal(a.ToStringMap())
	return string(json)
}

type AttributeValue interface {
	Type() AttributeValueType
	ToString() string
}

type stringValue struct {
	value string
}

func (v *stringValue) Type() AttributeValueType {
	return StringAttributeValueType
}

func (v stringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *stringValue) ToString() string {
	return v.value
}

type intValue struct {
	value int64
}

func (v *intValue) Type() AttributeValueType {
	return IntAttributeValueType
}

func (v *intValue) ToString() string {
	return strconv.FormatInt(v.value, 10)
}

func (v intValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

type boolValue struct {
	value bool
}

func (v *boolValue) Type() AttributeValueType {
	return BooleanAttributeValueType
}

func (v *boolValue) ToString() string {
	return strconv.FormatBool(v.value)
}

func (v boolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}
