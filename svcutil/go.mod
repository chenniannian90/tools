module github.com/chenniannian90/tools/svcutil

go 1.19

require (
	github.com/go-courier/courier v1.5.0
	github.com/go-courier/envconf v1.4.0
	github.com/go-courier/httptransport v1.22.2
	github.com/go-courier/logr v0.0.2
	github.com/go-courier/metax v1.3.0
	github.com/go-courier/statuserror v1.2.1
	github.com/go-courier/x v0.1.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.9.3
	go.opentelemetry.io/contrib/propagators v0.20.0
	go.opentelemetry.io/otel v0.20.0
	go.opentelemetry.io/otel/exporters/trace/zipkin v0.20.0
	go.opentelemetry.io/otel/sdk v0.20.0
	go.opentelemetry.io/otel/trace v0.20.0

)

require (
	github.com/fatih/color v1.15.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/openzipkin/zipkin-go v0.2.5 // indirect
	go.opentelemetry.io/otel/metric v0.20.0 // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.8.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	golang.org/x/tools v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/go-courier/httptransport => github.com/go-courier/httptransport v1.21.0
	github.com/go-courier/logr => github.com/go-courier/logr v0.0.2
)
