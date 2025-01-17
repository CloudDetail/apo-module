module github.com/CloudDetail/apo-module/slo/sdk

go 1.21

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.26.0
	github.com/CloudDetail/apo-module/slo/api v0.0.0-00000000000000-000000000000
	github.com/hashicorp/golang-lru/v2 v2.0.7
	github.com/olivere/elastic/v7 v7.0.32
	github.com/prometheus/client_golang v1.19.1
	github.com/prometheus/common v0.48.0
	github.com/robfig/cron/v3 v3.0.1
	go.uber.org/multierr v1.11.0
)

require (
	github.com/ClickHouse/ch-go v0.61.5 // indirect
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.7 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/paulmach/orb v0.11.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	go.opentelemetry.io/otel v1.26.0 // indirect
	go.opentelemetry.io/otel/trace v1.26.0 // indirect
	golang.org/x/sys v0.18.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/CloudDetail/apo-module/slo/api => ../api
