package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env           string             `mapstructure:"WISPMAIL_ENV"`
	LogLevel      string             `mapstructure:"LOG_LEVEL"`
	LogFormat     string             `mapstructure:"LOG_FORMAT"`
	API           APIConfig           `mapstructure:",squash"`
	SMTP          SMTPConfig          `mapstructure:",squash"`
	IMAP          IMAPConfig          `mapstructure:",squash"`
	POP3          POP3Config          `mapstructure:",squash"`
	JMAP          JMAPConfig          `mapstructure:",squash"`
	Database      DatabaseConfig      `mapstructure:",squash"`
	Redis         RedisConfig         `mapstructure:",squash"`
	Elasticsearch ElasticsearchConfig `mapstructure:",squash"`
	JWT           JWTConfig           `mapstructure:",squash"`
	S3            S3Config            `mapstructure:",squash"`
	DKIM          DKIMConfig          `mapstructure:",squash"`
	OTEL          OTELConfig          `mapstructure:",squash"`
	Metrics       MetricsConfig       `mapstructure:",squash"`
	RateLimit     RateLimitConfig     `mapstructure:",squash"`
	CORS          CORSConfig          `mapstructure:",squash"`
	Etcd          EtcdConfig          `mapstructure:",squash"`
	Cluster       ClusterConfig       `mapstructure:",squash"`
	GRPC          GRPCConfig          `mapstructure:",squash"`
}

type APIConfig struct {
	Host            string        `mapstructure:"API_HOST"`
	Port            int           `mapstructure:"API_PORT"`
	ReadTimeout     time.Duration `mapstructure:"API_READ_TIMEOUT"`
	WriteTimeout    time.Duration `mapstructure:"API_WRITE_TIMEOUT"`
	ShutdownTimeout time.Duration `mapstructure:"API_SHUTDOWN_TIMEOUT"`
}

type SMTPConfig struct {
	Host           string `mapstructure:"SMTP_HOST"`
	Port           int    `mapstructure:"SMTP_PORT"`
	TLSCert        string `mapstructure:"SMTP_TLS_CERT"`
	TLSKey         string `mapstructure:"SMTP_TLS_KEY"`
	MaxMessageSize int64  `mapstructure:"SMTP_MAX_MESSAGE_SIZE"`
}

type IMAPConfig struct {
	Host    string `mapstructure:"IMAP_HOST"`
	Port    int    `mapstructure:"IMAP_PORT"`
	TLSCert string `mapstructure:"IMAP_TLS_CERT"`
	TLSKey  string `mapstructure:"IMAP_TLS_KEY"`
}

type POP3Config struct {
	Host    string `mapstructure:"POP3_HOST"`
	Port    int    `mapstructure:"POP3_PORT"`
	TLSCert string `mapstructure:"POP3_TLS_CERT"`
	TLSKey  string `mapstructure:"POP3_TLS_KEY"`
}

type JMAPConfig struct {
	Host    string `mapstructure:"JMAP_HOST"`
	Port    int    `mapstructure:"JMAP_PORT"`
	TLSCert string `mapstructure:"JMAP_TLS_CERT"`
	TLSKey  string `mapstructure:"JMAP_TLS_KEY"`
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"DATABASE_URL"`
	MaxOpenConns    int           `mapstructure:"DATABASE_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DATABASE_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DATABASE_CONN_MAX_LIFETIME"`
	ConnMaxIdleTime time.Duration `mapstructure:"DATABASE_CONN_MAX_IDLE_TIME"`
}

type RedisConfig struct {
	URL          string        `mapstructure:"REDIS_URL"`
	PoolSize     int           `mapstructure:"REDIS_POOL_SIZE"`
	MinIdleConns int           `mapstructure:"REDIS_MIN_IDLE_CONNS"`
	DialTimeout  time.Duration `mapstructure:"REDIS_DIAL_TIMEOUT"`
	ReadTimeout  time.Duration `mapstructure:"REDIS_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"REDIS_WRITE_TIMEOUT"`
}

type ElasticsearchConfig struct {
	URL         string `mapstructure:"ELASTICSEARCH_URL"`
	IndexPrefix string `mapstructure:"ELASTICSEARCH_INDEX_PREFIX"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"JWT_SECRET"`
	AccessTTL  time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	RefreshTTL time.Duration `mapstructure:"JWT_REFRESH_TTL"`
	Issuer     string        `mapstructure:"JWT_ISSUER"`
}

type S3Config struct {
	Endpoint        string `mapstructure:"S3_ENDPOINT"`
	Region          string `mapstructure:"S3_REGION"`
	BucketName      string `mapstructure:"S3_BUCKET_NAME"`
	AccessKeyID     string `mapstructure:"S3_ACCESS_KEY_ID"`
	SecretAccessKey string `mapstructure:"S3_SECRET_ACCESS_KEY"`
	UseSSL          bool   `mapstructure:"S3_USE_SSL"`
	ForcePathStyle  bool   `mapstructure:"S3_FORCE_PATH_STYLE"`
}

type DKIMConfig struct {
	Selector       string `mapstructure:"DKIM_SELECTOR"`
	Domain         string `mapstructure:"DKIM_DOMAIN"`
	PrivateKeyPath string `mapstructure:"DKIM_PRIVATE_KEY_PATH"`
}

type OTELConfig struct {
	ExporterEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	ServiceName      string `mapstructure:"OTEL_SERVICE_NAME"`
	ServiceVersion   string `mapstructure:"OTEL_SERVICE_VERSION"`
}

type MetricsConfig struct {
	Path       string `mapstructure:"METRICS_PATH"`
	HealthPath string `mapstructure:"HEALTH_PATH"`
	ReadyPath  string `mapstructure:"READY_PATH"`
}

type RateLimitConfig struct {
	Enabled           bool `mapstructure:"RATE_LIMIT_ENABLED"`
	RequestsPerMinute int  `mapstructure:"RATE_LIMIT_REQUESTS_PER_MINUTE"`
	Burst             int  `mapstructure:"RATE_LIMIT_BURST"`
}

type CORSConfig struct {
	AllowedOrigins string `mapstructure:"CORS_ALLOWED_ORIGINS"`
	AllowedMethods string `mapstructure:"CORS_ALLOWED_METHODS"`
	AllowedHeaders string `mapstructure:"CORS_ALLOWED_HEADERS"`
}

type EtcdConfig struct {
	Endpoints   string        `mapstructure:"ETCD_ENDPOINTS"`
	DialTimeout time.Duration `mapstructure:"ETCD_DIAL_TIMEOUT"`
}

type ClusterConfig struct {
	Name          string `mapstructure:"CLUSTER_NAME"`
	NodeID        string `mapstructure:"CLUSTER_NODE_ID"`
	BindAddr      string `mapstructure:"CLUSTER_BIND_ADDR"`
	BindPort      int    `mapstructure:"CLUSTER_BIND_PORT"`
	AdvertiseAddr string `mapstructure:"CLUSTER_ADVERTISE_ADDR"`
	AdvertisePort int    `mapstructure:"CLUSTER_ADVERTISE_PORT"`
}

type GRPCConfig struct {
	Port              int  `mapstructure:"GRPC_PORT"`
	ReflectionEnabled bool `mapstructure:"GRPC_REFLECTION_ENABLED"`
}

func Load(configFile string) (*Config, error) {
	v := viper.New()

	if configFile != "" {
		v.SetConfigFile(configFile)
		v.SetConfigType("yaml")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}
	}

	v.SetEnvPrefix("WISPMAIL")
	v.AutomaticEnv()
	setDefaults(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

func MustLoad(configFile string) *Config {
	cfg, err := Load(configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("WISPMAIL_ENV", "development")
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")
	v.SetDefault("API_HOST", "0.0.0.0")
	v.SetDefault("API_PORT", 8080)
	v.SetDefault("API_READ_TIMEOUT", "30s")
	v.SetDefault("API_WRITE_TIMEOUT", "60s")
	v.SetDefault("API_SHUTDOWN_TIMEOUT", "30s")
	v.SetDefault("SMTP_HOST", "0.0.0.0")
	v.SetDefault("SMTP_PORT", 2525)
	v.SetDefault("SMTP_MAX_MESSAGE_SIZE", 26214400)
	v.SetDefault("IMAP_HOST", "0.0.0.0")
	v.SetDefault("IMAP_PORT", 993)
	v.SetDefault("POP3_HOST", "0.0.0.0")
	v.SetDefault("POP3_PORT", 995)
	v.SetDefault("JMAP_HOST", "0.0.0.0")
	v.SetDefault("JMAP_PORT", 443)
	v.SetDefault("DATABASE_MAX_OPEN_CONNS", 25)
	v.SetDefault("DATABASE_MAX_IDLE_CONNS", 10)
	v.SetDefault("DATABASE_CONN_MAX_LIFETIME", "30m")
	v.SetDefault("DATABASE_CONN_MAX_IDLE_TIME", "5m")
	v.SetDefault("REDIS_POOL_SIZE", 20)
	v.SetDefault("REDIS_MIN_IDLE_CONNS", 5)
	v.SetDefault("REDIS_DIAL_TIMEOUT", "5s")
	v.SetDefault("REDIS_READ_TIMEOUT", "3s")
	v.SetDefault("REDIS_WRITE_TIMEOUT", "3s")
	v.SetDefault("ELASTICSEARCH_INDEX_PREFIX", "wispmail")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "7d")
	v.SetDefault("JWT_ISSUER", "wispmail")
	v.SetDefault("S3_REGION", "us-east-1")
	v.SetDefault("METRICS_PATH", "/metrics")
	v.SetDefault("HEALTH_PATH", "/healthz")
	v.SetDefault("READY_PATH", "/readyz")
	v.SetDefault("RATE_LIMIT_ENABLED", true)
	v.SetDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 100)
	v.SetDefault("RATE_LIMIT_BURST", 20)
	v.SetDefault("CORS_ALLOWED_METHODS", "GET,POST,PUT,DELETE,OPTIONS")
	v.SetDefault("CORS_ALLOWED_HEADERS", "Authorization,Content-Type,X-Request-ID")
	v.SetDefault("ETCD_DIAL_TIMEOUT", "5s")
	v.SetDefault("CLUSTER_NAME", "wispmail")
	v.SetDefault("CLUSTER_NODE_ID", "node-1")
	v.SetDefault("CLUSTER_BIND_PORT", 7946)
	v.SetDefault("CLUSTER_ADVERTISE_PORT", 7946)
	v.SetDefault("GRPC_PORT", 9090)
	v.SetDefault("GRPC_REFLECTION_ENABLED", false)
}