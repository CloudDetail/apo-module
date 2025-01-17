package model

import (
	"sort"
	"strconv"
	"strings"
)

type Logs struct {
	logMap map[uint64]*Log
}

func newLogs() *Logs {
	return &Logs{
		logMap: make(map[uint64]*Log),
	}
}

func (logs *Logs) addLog(tid uint64, name string, logText string) {
	log, exist := logs.logMap[tid]
	if !exist {
		log = &Log{
			ThreadId:   tid,
			ThreadName: name,
			Logs:       make([]string, 0),
		}
		logs.logMap[tid] = log
	}
	logList := SplitLogs(logText)
	if len(logList) > 0 {
		log.Logs = append(log.Logs, logList...)
	}
}

func (logs *Logs) MergeLog(other *Logs) *Logs {
	for k, v := range other.logMap {
		logs.logMap[k] = v
	}
	return logs
}

func (logs *Logs) GetSortedLogsByThreadName() []*Log {
	sortedLogs := make([]*Log, 0, len(logs.logMap))
	for _, v := range logs.logMap {
		sortedLogs = append(sortedLogs, v)
	}
	sort.Sort(byThreadName(sortedLogs))
	return sortedLogs
}

func SplitLogs(logText string) []string {
	logs := make([]string, 0)
	size := len([]rune(logText))
	var tempLog = logText
	for startIndex := 0; startIndex < size; {
		tempLog = tempLog[startIndex:]
		lenIndex := strings.Index(tempLog, "@")
		if lenIndex == -1 {
			return logs
		}
		logSize, err := strconv.Atoi(tempLog[:lenIndex])
		if err != nil {
			return logs
		}

		endIndex := lenIndex + 1 + logSize
		if endIndex > size || tempLog[endIndex] != '|' {
			return logs
		}
		if logSize > 0 {
			singleLog := tempLog[lenIndex+1 : endIndex]
			logs = append(logs, strings.Split(singleLog, "<br>")...)
		}
		startIndex = endIndex + 1
	}
	return logs
}

type Log struct {
	ThreadId   uint64   `json:"id"`
	ThreadName string   `json:"name"`
	Logs       []string `json:"values"`
}

type byThreadName []*Log

func (m byThreadName) Len() int {
	return len(m)
}

func (m byThreadName) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m byThreadName) Less(i, j int) bool {
	if m[i].ThreadName == m[j].ThreadName {
		return m[i].ThreadId < m[j].ThreadId
	}
	return m[i].ThreadName < m[j].ThreadName
}
