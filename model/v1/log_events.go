package model

type LogEvents struct {
	StartTime uint64
	EndTime   uint64
	Logs      map[uint64][]*LogEvent
}

func NewLogEvents(startTime uint64, endTime uint64) *LogEvents {
	return &LogEvents{
		StartTime: startTime,
		EndTime:   endTime,
		Logs:      make(map[uint64][]*LogEvent, 0),
	}
}

type LogEvent struct {
	Pid         uint64 `json:"pid"`
	Tid         uint64 `json:"tid"`
	ThreadName  string `json:"threadName"`
	ContainerId string `json:"container_id"`
	NodeName    string `json:"node_name"`
	NodeIp      string `json:"node_ip"`
	Logs        string `json:"logs"`
}
