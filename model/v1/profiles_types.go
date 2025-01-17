package model

type CPUType uint8

const (
	CPUType_ON           CPUType = 0
	CPUType_FILE         CPUType = 1
	CPUType_NET          CPUType = 2
	CPUType_FUTEX        CPUType = 3
	CPUType_IDLE         CPUType = 4
	CPUType_OTHER        CPUType = 5
	CPUType_EPOLL        CPUType = 6
	CPUType_RUNQ         CPUType = 7
	CPUTYPE_MAX          CPUType = 8
	CPUTYPE_MISSING_DATA CPUType = 9
)

type AggregatedTime struct {
	// on, file, net, futex, idle, other, epoll
	Times [CPUTYPE_MAX]uint64
}

type CpuEvent struct {
	StartTime   uint64    `json:"startTime"`
	EndTime     uint64    `json:"endTime"`
	TypeSpecs   []uint64  `json:"typeSpecs"`
	RunqLatency []uint64  `json:"runqLatency"`
	TimeType    []CPUType `json:"timeType"`
	OnInfo      string    `json:"onInfo"`
	OffInfo     string    `json:"offInfo"`
	Log         string    `json:"log"`
	Stack       string    `json:"stack"`
}

type TransactionIdEvent struct {
	Timestamp uint64 `json:"timestamp"`
	TraceId   string `json:"traceId"`
	IsEntry   uint32 `json:"isEntry"`
}

type JavaFutexEvent struct {
	StartTime uint64 `json:"startTime"`
	EndTime   uint64 `json:"endTime"`
	DataVal   string `json:"dataValue"`
}

type CameraEventGroup struct {
	Timestamp uint64           `json:"timestamp"`
	Labels    CameraEventLabel `json:"labels"`
}

type CameraEventLabel struct {
	ContentKey      string `json:"content_key"`
	CpuEvents       string `json:"cpuEvents"`
	JavaFutexEvents string `json:"javaFutexEvents"`
	Tid             uint64 `json:"tid"`
	ThreadName      string `json:"threadName"`
	OffsetTs        int64  `json:"offset_ts"`
}
