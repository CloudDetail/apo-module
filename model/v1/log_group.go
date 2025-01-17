package model

type CameraLogGroup struct {
	Name      string         `json:"name"`
	Timestamp int64          `json:"timestamp"`
	Labels    CameraLogLabel `json:"labels"`
}
type CameraLogLabel struct {
	ContainerId string `json:"container_id"`
	NodeName    string `json:"node_name"`
	NodeIp      string `json:"node_ip"`
	IsSent      int    `json:"isSent"`
	Pid         uint64 `json:"pid"`
	ThreadName  string `json:"threadName"`
	Tid         uint64 `json:"tid"`
	Logs        string `json:"logs"`
}
