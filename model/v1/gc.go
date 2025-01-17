package model

import "time"

type Gc struct {
	YgcStartTime uint64 `json:"ygc_last_entry_time"`
	YgcDuration  uint64 `json:"ygc_span"`
	YgcCount     int    `json:"ygc"`
	LastYgcCount int    `json:"last_ygc"`
	FgcStartTime uint64 `json:"fgc_last_entry_time"`
	FgcDuration  uint64 `json:"fgc_span"`
	FgcCount     int    `json:"fgc"`
	LastFgcCount int    `json:"last_fgc"`
}

func (gc *Gc) Match(start uint64, end uint64) bool {
	startMs := start / 1000000
	endMs := end / 1000000
	if gc.IsYgc() && startMs <= (gc.YgcStartTime+gc.YgcDuration/1e6) && endMs >= gc.YgcStartTime {
		return true
	}
	if gc.IsFgc() && startMs <= (gc.FgcStartTime+gc.FgcDuration/1e6) && endMs >= gc.FgcStartTime {
		return true
	}
	return false
}

func (gc *Gc) IsYgc() bool {
	return gc.YgcCount > gc.LastYgcCount
}

func (gc *Gc) IsFgc() bool {
	return gc.FgcCount > gc.LastFgcCount
}

func (gc *Gc) GetYGcStartTimeStr() string {
	return time.UnixMilli(int64(gc.YgcStartTime)).Format("2006-01-02 15:04:05.000")
}

func (gc *Gc) GetFGcStartTimeStr() string {
	return time.UnixMilli(int64(gc.FgcStartTime)).Format("2006-01-02 15:04:05.000")
}

func (gc *Gc) GetYGcDuration() uint64 {
	return gc.YgcDuration
}

func (gc *Gc) GetFGcDuration() uint64 {
	return gc.FgcDuration
}

func (gc *Gc) GetYGcType() string {
	return "Young GC"
}

func (gc *Gc) GetFGcType() string {
	return "Full GC"
}

func (gc *Gc) GetLastYGcCount() int {
	return gc.LastYgcCount
}

func (gc *Gc) GetYGcCount() int {
	return gc.YgcCount
}

func (gc *Gc) GetLastFGcCount() int {
	return gc.LastFgcCount
}

func (gc *Gc) GetFGcCount() int {
	return gc.FgcCount
}
