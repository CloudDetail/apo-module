package model

import (
	"fmt"

	conventions "go.opentelemetry.io/collector/semconv/v1.9.0"
)

type ApmClientCall struct {
	ClientStartTime  uint64            `json:"client_start_time"`
	ClientEndTime    uint64            `json:"client_end_time"`
	ClientName       string            `json:"client_name"`
	ClientSpanId     string            `json:"client_spanid"`
	ClientAttributes map[string]string `json:"client_attributes"`
	ServerDuration   uint64            `json:"server_duration"`
	ServerName       string            `json:"server_name"`

	ClientOriginalSpanId string `json:"-"`
}

func (clientCall *ApmClientCall) GetClentInfo() (string, string) {
	if len(clientCall.ClientName) == 0 || len(clientCall.ClientAttributes) == 0 {
		return "unknown", "unknown"
	}
	if url, exist := clientCall.ClientAttributes[conventions.AttributeHTTPURL]; exist {
		return "http", url
	}
	if statement, exist := clientCall.ClientAttributes[conventions.AttributeDBStatement]; exist {
		dbType := clientCall.ClientAttributes[conventions.AttributeDBSystem]
		if len(statement) == 0 {
			return dbType, clientCall.ClientName
		}
		return dbType, statement
	}
	if broker, exist := clientCall.ClientAttributes["mq.broker"]; exist {
		return clientCall.ClientName, fmt.Sprintf("Broker-%s", broker)
	}
	return "unknown", clientCall.ClientName
}

type ReqKind int

const (
	UnknownReqKind ReqKind = 0
	HTTPReqKind    ReqKind = 1
	SQLReqKind     ReqKind = 2
	MQReqKind      ReqKind = 3
)

type ClientInfo struct {
	ReqKind    ReqKind
	ReqType    string
	ReqContent string
}

func (clientCall *ApmClientCall) ClientInfo() *ClientInfo {
	if len(clientCall.ClientName) == 0 || len(clientCall.ClientAttributes) == 0 {
		return &ClientInfo{
			ReqKind:    UnknownReqKind,
			ReqType:    "unknown",
			ReqContent: "unknown",
		}
	}
	if url, exist := clientCall.ClientAttributes[conventions.AttributeHTTPURL]; exist {
		return &ClientInfo{
			ReqKind:    HTTPReqKind,
			ReqType:    "http",
			ReqContent: url,
		}
	}
	if url, exist := clientCall.ClientAttributes["url.full"]; exist {
		return &ClientInfo{
			ReqKind:    HTTPReqKind,
			ReqType:    "http",
			ReqContent: url,
		}
	}
	if statement, exist := clientCall.ClientAttributes[conventions.AttributeDBStatement]; exist {
		dbType := clientCall.ClientAttributes[conventions.AttributeDBSystem]
		clientInfo := &ClientInfo{
			ReqKind:    SQLReqKind,
			ReqType:    dbType,
			ReqContent: statement,
		}
		if len(statement) == 0 {
			clientInfo.ReqContent = clientCall.ClientName
		}
		return clientInfo
	}
	if broker, exist := clientCall.ClientAttributes["mq.broker"]; exist {
		return &ClientInfo{
			ReqKind:    MQReqKind,
			ReqType:    clientCall.ClientName,
			ReqContent: fmt.Sprintf("Broker-%s", broker),
		}

	}
	return &ClientInfo{
		ReqKind:    UnknownReqKind,
		ReqType:    "unknown",
		ReqContent: clientCall.ClientName,
	}
}
