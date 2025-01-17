package model

import (
	"fmt"
	"strings"
)

var CPUTypes = [CPUTYPE_MAX]string{
	CPUType_ON:    "cpu",
	CPUType_FILE:  "file",
	CPUType_NET:   "net",
	CPUType_FUTEX: "futex",
	CPUType_IDLE:  "idle",
	CPUType_OTHER: "other",
	CPUType_EPOLL: "epoll",
	CPUType_RUNQ:  "runq",
}

type Profiles struct {
	StartTime  uint64
	EndTime    uint64
	Tid        uint64
	ThreadName string
	AggTime    AggregatedTime
	CpuEvents  []*CpuEvent
	Futexs     []*JavaFutexEvent
	OffsetTs   []int64
}

func NewProfiles(startTime uint64, endTime uint64) *Profiles {
	return &Profiles{
		StartTime: startTime,
		EndTime:   endTime,
		AggTime:   AggregatedTime{},
		CpuEvents: make([]*CpuEvent, 0),
		Futexs:    make([]*JavaFutexEvent, 0),
	}
}

func (p *Profiles) AddCpuEvents(events []*CpuEvent) {
	var lastCpuEndTime uint64 = 0
	if len(p.CpuEvents) > 0 {
		lastCpuEndTime = p.CpuEvents[len(p.CpuEvents)-1].EndTime
	}
	for _, event := range events {
		if event.EndTime <= lastCpuEndTime {
			continue
		}
		if p.isMatch(event.StartTime, event.EndTime, 1) {
			p.CpuEvents = append(p.CpuEvents, event)
		}
	}
}

func (p *Profiles) CalcProfileEventMetrics() uint64 {
	var currentTime uint64 = 0
	sumTime := uint64(0)
	for _, event := range p.CpuEvents {
		currentTime = event.StartTime
		for i := 0; i < len(event.TypeSpecs); i++ {
			onoffTime, runqTime := p.getOnOffRunqTime(event, i, currentTime)
			p.AggTime.Times[event.TimeType[i]] += onoffTime
			p.AggTime.Times[CPUType_RUNQ] += runqTime
			sumTime = sumTime + onoffTime + runqTime
			currentTime += event.TypeSpecs[i]
		}
	}
	return sumTime
}

func (p *Profiles) getOnOffRunqTime(event *CpuEvent, i int, currentTime uint64) (onoffTime uint64, runqTime uint64) {
	onoffTime = event.TypeSpecs[i]
	if p.StartTime-onoffTime > currentTime || p.EndTime < currentTime {
		return 0, 0
	}

	if currentTime < p.StartTime {
		onoffTime = onoffTime + currentTime - p.StartTime
	} else if currentTime+onoffTime > p.EndTime {
		onoffTime = p.EndTime - currentTime
	}

	if event.TimeType[i] > 0 {
		runqTime = event.RunqLatency[i/2] * 1000 // us -> ns
		if runqTime > 0 {
			if onoffTime < runqTime {
				runqTime = onoffTime
				onoffTime = 0
			} else {
				// OffData Minus Runq
				onoffTime -= runqTime
			}
		}
	}
	return onoffTime, runqTime
}

func (p *Profiles) isMatch(start uint64, end uint64, correctMilli uint64) bool {
	return p.StartTime/1000000 <= (end/1000000)+correctMilli || (p.EndTime/1000000)+correctMilli >= start
}

func (p *Profiles) AddFutexEvents(events []*JavaFutexEvent) {
	for _, event := range events {
		if event.StartTime >= p.StartTime && event.EndTime <= p.EndTime {
			p.Futexs = append(p.Futexs, event)
		}
	}
}

func (p *Profiles) GetLogs() *Logs {
	logs := newLogs()
	for _, event := range p.CpuEvents {
		if len(event.Log) > 0 {
			logs.addLog(p.Tid, p.ThreadName, event.Log)
		}
	}
	return logs
}

func (profiles *Profiles) ToString() string {
	var text strings.Builder
	text.WriteString(fmt.Sprintf("StartTime: %d, EndTime: %d, ", profiles.StartTime, profiles.EndTime))
	text.WriteString(fmt.Sprintf("AggTime: %s, CpuEvents: [\n", structToString(profiles.AggTime)))
	for i, profile := range profiles.CpuEvents {
		if i > 0 {
			text.WriteString(",\n")
		}
		text.WriteString(fmt.Sprintf("    {%s}", structToString(profile)))
	}
	text.WriteString("\n]")
	return text.String()
}
