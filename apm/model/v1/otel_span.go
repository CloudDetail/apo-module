package model

import (
	"fmt"

	cmodel "github.com/CloudDetail/apo-module/model/v1"
)

type OtelSpan struct {
	StartTime   uint64              `json:"startTime"` // ns
	Duration    uint64              `json:"duration"`  // ns
	ServiceName string              `json:"serviceName,omitempty"`
	Name        string              `json:"name"`
	SpanId      string              `json:"spanId,omitempty"`
	PSpanId     string              `json:"pSpanId,omitempty"`
	NextSpanId  string              `json:"nextSpanId,omitempty"`
	Kind        OtelSpanKind        `json:"kind"`
	Code        OtelStatusCode      `json:"code"`
	NotSampled  bool                `json:"-"`
	Attributes  map[string]string   `json:"attributes,omitempty"`
	Exceptions  []*cmodel.Exception `json:"exceptions,omitempty"`
}

func NewOtelSpan() *OtelSpan {
	return &OtelSpan{
		Attributes: make(map[string]string, 0),
	}
}

func (span *OtelSpan) SetStartTime(startTime uint64) {
	span.StartTime = startTime
}

func (span *OtelSpan) SetDuration(duration uint64) {
	span.Duration = duration
}

func (span *OtelSpan) SetName(name string) {
	span.Name = name
}

func (span *OtelSpan) SetServiceName(serviceName string) {
	span.ServiceName = serviceName
}

func (span *OtelSpan) SetSpanId(spanId string) {
	span.SpanId = spanId
}

func (span *OtelSpan) SetParentSpanId(pSpanId string) {
	span.PSpanId = pSpanId
}

func (span *OtelSpan) SetKind(kind OtelSpanKind) {
	span.Kind = kind
}

func (span *OtelSpan) SetCode(code OtelStatusCode) {
	span.Code = code
}

func (span *OtelSpan) IsError() bool {
	return span.Code == StatusCodeError
}

func (span *OtelSpan) AddAttribute(key string, value string) {
	span.Attributes[key] = value
}

func (span *OtelSpan) AddException(timestamp uint64, name string, message string, stack string) {
	if span.Exceptions == nil {
		span.Exceptions = make([]*cmodel.Exception, 0)
	}
	span.Exceptions = append(span.Exceptions, cmodel.NewOtelException(timestamp, name, message, stack))
}

func (span *OtelSpan) GetEndTime() uint64 {
	return span.StartTime + span.Duration
}

func (span *OtelSpan) SetOriginalSpanId(apmType string, spanId string) {
	span.Attributes[AttributeApmSpanType] = apmType
	span.Attributes[AttributeApmOriginalSpanId] = spanId
}

func (span *OtelSpan) OriginalSpanId() string {
	return span.Attributes[AttributeApmOriginalSpanId]
}

func (span *OtelSpan) ApmType() string {
	return span.Attributes[AttributeApmSpanType]
}

func (span *OtelSpan) GetHttpMethod() string {
	// 1.x http.method
	if httpMethod := span.Attributes[AttributeHttpMethod]; httpMethod != "" {
		return httpMethod
	}
	// 2.x http.request.method
	return span.Attributes[AttributeHttpRequestMethod]
}

func (span *OtelSpan) GetHttpDetail() string {
	// 1.x http.url
	if httpMethod := span.Attributes[AttributeHTTPURL]; httpMethod != "" {
		return httpMethod
	}
	// 2.x url.full
	return span.Attributes[AttributeURLFULL]
}

func (span *OtelSpan) GetPeer(defaultValue string) string {
	// 1.x - redis、grpc、rabbitmq
	if netSockPeerAddr, addrFound := span.Attributes[AttributeNetSockPeerAddr]; addrFound {
		if netSockPeerPort, portFound := span.Attributes[AttributeNetSockPeerPort]; portFound {
			return fmt.Sprintf("%s:%s", netSockPeerAddr, netSockPeerPort)
		} else {
			return netSockPeerAddr
		}
	}
	// 2.x - redis、grpc、rabbitmq
	if networkPeerAddress, addrFound := span.Attributes[AttributeNetworkPeerAddress]; addrFound {
		if networkPeerPort, portFound := span.Attributes[AttributeNetworkPeerPort]; portFound {
			return fmt.Sprintf("%s:%s", networkPeerAddress, networkPeerPort)
		} else {
			return networkPeerAddress
		}
	}

	// 1.x - httpclient、db、dubbo
	if peerName, peerFound := span.Attributes[AttributeNetPeerName]; peerFound {
		if peerPort, peerPortFound := span.Attributes[AttributeNetPeerPort]; peerPortFound {
			return fmt.Sprintf("%s:%s", peerName, peerPort)
		} else {
			return peerName
		}
	}
	// 2.x - httpclient、db、dubbo
	if serverAddress, serverFound := span.Attributes[AttributeServerAddress]; serverFound {
		if serverPort, serverPortFound := span.Attributes[AttributeServerPort]; serverPortFound {
			return fmt.Sprintf("%s:%s", serverAddress, serverPort)
		} else {
			return serverAddress
		}
	}

	return defaultValue
}

func (span *OtelSpan) GetMessageDestination(defaultValue string) string {
	// 1.x messaging.destination
	if messageDest, found := span.Attributes[AttributeMessageDestination]; found {
		return messageDest
	}
	// 2.x messaging.destination.name
	if messageDestName, found := span.Attributes[AttributeMessageDestinationName]; found {
		return messageDestName
	}
	return defaultValue
}

func (span *OtelSpan) GetRpcDetail(defaultValue string) string {
	rpcService := span.Attributes[AttributeRpcService]
	rpcMethod := span.Attributes[AttributeRpcMethod]
	if rpcService != "" && rpcMethod != "" {
		return fmt.Sprintf("%s/%s", rpcService, span.Attributes[AttributeRpcMethod])
	}

	// Skywalking [full.url]
	if fullUrl := span.Attributes[AttributeURLFULL]; fullUrl != "" {
		return fullUrl
	}
	return defaultValue
}
