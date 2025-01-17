module github.com/CloudDetail/apo-module/apm/client

go 1.21

require (
	github.com/CloudDetail/apo-module/apm/model v0.0.0-00000000000000-000000000000
	github.com/CloudDetail/apo-module/model v0.0.0-00000000000000-000000000000
	github.com/xwb1989/sqlparser v0.0.0-20180606152119-120387863bf2
)

require go.opentelemetry.io/collector/semconv v0.97.0 // indirect

replace (
	github.com/CloudDetail/apo-module/apm/model => ../model
	github.com/CloudDetail/apo-module/model => ../../model
)
