package model

type OtelSpanKind int32

const (
	SpanKindUnspecified = OtelSpanKind(0)
	SpanKindInternal    = OtelSpanKind(1)
	SpanKindServer      = OtelSpanKind(2)
	SpanKindClient      = OtelSpanKind(3)
	SpanKindProducer    = OtelSpanKind(4)
	SpanKindConsumer    = OtelSpanKind(5)
)

func (sk OtelSpanKind) String() string {
	switch sk {
	case SpanKindUnspecified:
		return "Unspecified"
	case SpanKindInternal:
		return "Internal"
	case SpanKindServer:
		return "Server"
	case SpanKindClient:
		return "Client"
	case SpanKindProducer:
		return "Producer"
	case SpanKindConsumer:
		return "Consumer"
	}
	return ""
}

type OtelStatusCode int32

const (
	StatusCodeUnset = OtelStatusCode(0)
	StatusCodeOk    = OtelStatusCode(1)
	StatusCodeError = OtelStatusCode(2)
)

func (sc OtelStatusCode) String() string {
	switch sc {
	case StatusCodeUnset:
		return "Unset"
	case StatusCodeOk:
		return "Ok"
	case StatusCodeError:
		return "Error"
	}
	return ""
}

func (kind OtelSpanKind) IsExit() bool {
	return kind == SpanKindClient || kind == SpanKindProducer
}

func (kind OtelSpanKind) IsEntry() bool {
	return kind == SpanKindServer || kind == SpanKindConsumer
}
