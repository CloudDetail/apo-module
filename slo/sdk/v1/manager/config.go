package manager

import (
	"time"

	"github.com/CloudDetail/apo-module/slo/sdk/v1/clickhouse"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/config"
	"github.com/CloudDetail/apo-module/slo/sdk/v1/elasticsearch"
)

type SLOManagerConfig struct {
	Enable                  bool `mapstructure:"enable"`
	EnableDebugInfo         bool `mapstructure:"debug"`
	EnableStorage           bool `mapstructure:"enable_slo_record"`
	EnableGenerateSLORecord bool `mapstructure:"enable_generate_slo_record"`

	CenterServer config.CenterServerConfig `mapstructure:"center_server"`
	Checker      CheckerConfig             `mapstructure:"checker"`
	Storage      StorageConfig             `mapstructure:"storage"`
}

func DefaultSLOConfig() *SLOManagerConfig {
	return &SLOManagerConfig{
		Enable:                  true,
		EnableDebugInfo:         false,
		EnableStorage:           true,
		EnableGenerateSLORecord: true,
		Storage: StorageConfig{
			StorageType: "clickhouse",
			Clickhouse: clickhouse.ClickhouseConfig{
				Authentication: clickhouse.Authentication{
					PlainText: &clickhouse.PlainTextConfig{
						Database: "originx",
					},
				},
				Table:            "slo_record",
				Compression:      "lz4",
				MaxExecutionTime: 60,
				DialTimeout:      time.Duration(10) * time.Second,
				MaxOpenConns:     5,
				MaxIdleConns:     5,
				ConnMaxLifetime:  time.Duration(10) * time.Minute,
				BlockBufferSize:  10,
				BufferNumLayers:  16,
				BufferMinTime:    10,
				BufferMaxTime:    100,
				BufferMinRows:    10000,
				BufferMaxRows:    1000000,
				BufferMinBytes:   10000000,
				BufferMaxBytes:   100000000,
			},
		},
	}
}

type CheckerConfig struct {
	PrometheusAddr string `mapstructure:"prometheus_addr"`
	PQLType        string `mapstructure:"pql_type"`
}

type StorageConfig struct {
	StorageType   string                            `mapstructure:"storage_type"`
	Elasticsearch elasticsearch.ElasticsearchConfig `mapstructure:"elasticsearch"`
	Clickhouse    clickhouse.ClickhouseConfig       `mapstructure:"clickhouse"`
}
