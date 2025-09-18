module github.com/mexxo-dvp/sentinel-bot

go 1.24.5

require (
	github.com/hirokisan/zerodriver v1.10.2
	github.com/spf13/cobra v1.9.1
	go.opentelemetry.io/otel v1.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.26.0
	go.opentelemetry.io/otel/metric v1.26.0
	go.opentelemetry.io/otel/sdk v1.26.0
	go.opentelemetry.io/otel/semconv/v1.24.0 v1.24.0
	gopkg.in/telebot.v3 v3.3.8
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	google.golang.org/grpc v1.63.2 // indirect
)
