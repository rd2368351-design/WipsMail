module github.com/wispmail/wispmail

go 1.26.5

require (
	github.com/aws/aws-sdk-go-v2 v1.36.3
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67
	github.com/aws/aws-sdk-go-v2/service/s3 v1.85.0
	github.com/emersion/go-smtp v0.21.3
	github.com/go-playground/validator/v10 v10.24.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/golang-migrate/migrate/v4 v4.18.2
	github.com/google/uuid v1.6.0
	github.com/hibiken/asynq v0.25.1
	github.com/jackc/pgx/v5 v5.7.4
	github.com/oklog/ulid/v2 v2.1.0
	github.com/prometheus/client_golang v1.21.1
	github.com/redis/go-redis/v9 v9.7.3
	github.com/rs/zerolog v1.34.0
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	go.etcd.io/etcd/client/v3 v3.5.19
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.60.0
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.35.0
	go.opentelemetry.io/otel/sdk v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
	go.uber.org/mock v0.5.0
	golang.org/x/crypto v0.38.0
	golang.org/x/sync v0.15.0
	google.golang.org/grpc v1.73.0
	google.golang.org/protobuf v1.36.6
	gopkg.in/yaml.v3 v3.0.1
)