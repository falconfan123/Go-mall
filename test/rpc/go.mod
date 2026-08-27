module github.com/falconfan123/Go-mall/test/rpc

go 1.25.0

replace github.com/falconfan123/Go-mall/common => ../../common

replace github.com/falconfan123/Go-mall/dal => ../../dal

replace github.com/falconfan123/Go-mall/services/activity => ../../services/activity

replace github.com/falconfan123/Go-mall/services/audit => ../../services/audit

replace github.com/falconfan123/Go-mall/services/auths => ../../services/auths

replace github.com/falconfan123/Go-mall/services/carts => ../../services/carts

replace github.com/falconfan123/Go-mall/services/checkout => ../../services/checkout

replace github.com/falconfan123/Go-mall/services/coupons => ../../services/coupons

replace github.com/falconfan123/Go-mall/services/inventory => ../../services/inventory

replace github.com/falconfan123/Go-mall/services/order => ../../services/order

replace github.com/falconfan123/Go-mall/services/payment => ../../services/payment

replace github.com/falconfan123/Go-mall/services/product => ../../services/product

replace github.com/falconfan123/Go-mall/services/users => ../../services/users

require (
	github.com/falconfan123/Go-mall/common v0.0.0-20260312153719-88b43b07ae7d
	github.com/falconfan123/Go-mall/dal v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/audit v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/auths v0.0.0-20260526005452-837c41f1788a
	github.com/falconfan123/Go-mall/services/carts v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/checkout v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/coupons v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/inventory v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/order v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/payment v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/product v0.0.0-00010101000000-000000000000
	github.com/falconfan123/Go-mall/services/users v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/olivere/elastic/v7 v7.0.32
	github.com/stretchr/testify v1.11.1
	github.com/zeromicro/go-zero v1.10.1
	google.golang.org/grpc v1.82.1
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-sql-driver/mysql v1.9.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.28.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pelletier/go-toml/v2 v2.3.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/spaolacci/murmur3 v1.1.0 // indirect
	github.com/titanous/json5 v1.0.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.67.0 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.43.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/text v0.39.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
