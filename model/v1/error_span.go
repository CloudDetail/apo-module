package model

type ErrorSpan struct {
	Name       string            `json:"name"`
	StartTime  uint64            `json:"startTime"`
	TotalTime  uint64            `json:"totalTime"`
	Attributes map[string]string `json:"attributes"`
	Exceptions []*Exception      `json:"exceptions"`
}

func NewErrorSpan(name string, startTime uint64, duration uint64) *ErrorSpan {
	return &ErrorSpan{
		Name:       name,
		StartTime:  startTime,
		TotalTime:  duration,
		Attributes: make(map[string]string),
		Exceptions: make([]*Exception, 0),
	}
}

func (span *ErrorSpan) AddAttribute(key string, value string) {
	span.Attributes[key] = value
}

type Exception struct {
	Timestamp uint64 `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	Stack     string `json:"stack"`
}

func NewOtelException(timestamp uint64, name string, message string, stack string) *Exception {
	return &Exception{
		Timestamp: timestamp,
		Type:      name,
		Message:   message,
		Stack:     stack,
	}
}
