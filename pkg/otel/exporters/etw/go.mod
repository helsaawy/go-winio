module github.com/Microsoft/go-winio/pkg/otel/exporters/etw

go 1.17

require (
	github.com/Microsoft/go-winio v0.6.0
	go.opentelemetry.io/otel/sdk v1.14.0
)

require (
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	go.opentelemetry.io/otel v1.14.0 // indirect
	go.opentelemetry.io/otel/trace v1.14.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)

replace github.com/Microsoft/go-winio => ../../../..
