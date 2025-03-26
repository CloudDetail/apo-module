package model

type AgentEvent struct {
	Timestamp uint64            `json:"timestamp"`
	Name      string            `json:"name"`
	Pid       uint32            `json:"pid"`
	Labels    map[string]string `json:"labels"`
	Status    bool              `json:"status"`
}
