package model

type SLOType string

const (
	SLO_SUCCESS_RATE_TYPE SLOType = "SuccessRate"
	SLO_LATENCY_P90_TYPE  SLOType = "LatencyP90"
	SLO_LATENCY_P95_TYPE  SLOType = "LatencyP95"
	SLO_LATENCY_P99_TYPE  SLOType = "LatencyP99"
)

type SLOStatus string

const (
	Achieved    SLOStatus = "Achieved"
	NotAchieved SLOStatus = "NotAchieved"
	Unknown     SLOStatus = "Unknown"
)

type ExpectedSource string

const (
	LastHourExpectedSource ExpectedSource = "last1h"
	YesterdayExpectSource  ExpectedSource = "yesterday"
	ConstantExpectSource   ExpectedSource = "constant"
	DefaultExpectSource    ExpectedSource = "default"
)

type SLOConfig struct {
	Type          SLOType        `json:"type" db:"slo_type"`
	Multiple      float64        `json:"multiple"`
	ExpectedValue float64        `json:"expectedValue" db:"expected_value"`
	Source        ExpectedSource `json:"source"`
}

type SLO struct {
	SLOConfig    *SLOConfig `json:"sloConfig"`
	CurrentValue float64    `json:"currentValue"`
	Status       SLOStatus  `json:"status"`
}

func GetLatencyPercentileByType(sloType SLOType) float64 {
	switch sloType {
	case SLO_LATENCY_P90_TYPE:
		return 0.9
	case SLO_LATENCY_P95_TYPE:
		return 0.95
	case SLO_LATENCY_P99_TYPE:
		return 0.99
	default:
		panic("not a LatencyPercentile type")
	}
}

func IsLatencyPercentileSLOType(sloType SLOType) bool {
	switch sloType {
	case SLO_LATENCY_P90_TYPE, SLO_LATENCY_P95_TYPE, SLO_LATENCY_P99_TYPE:
		return true
	default:
		return false
	}
}

type SLOEntryKey struct {
	EntryURI string
}

type SLOEntryKeyTemp struct {
	EntryURI     string
	EntryService string
}

type SLOEntryInfo struct {
	KeyRef *SLOEntryKey `json:"-"`
	Alias  string       `json:"alias"`
}

type SLOTarget struct {
	InfoRef    *SLOEntryInfo `json:"-"`
	SLOConfigs []SLOConfig   `json:"sloConfgs"`
}
