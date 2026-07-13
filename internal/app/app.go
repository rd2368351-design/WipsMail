package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/wispmail/wispmail/config"
	"github.com/wispmail/wispmail/internal/api"
	"github.com/wispmail/wispmail/internal/auth/apikey"
	"github.com/wispmail/wispmail/internal/auth/jwt"
	"github.com/wispmail/wispmail/internal/auth/rbac"
	"github.com/wispmail/wispmail/internal/dkim"
	"github.com/wispmail/wispmail/internal/dmarc"
	"github.com/wispmail/wispmail/internal/imap"
	"github.com/wispmail/wispmail/internal/jmap"
	"github.com/wispmail/wispmail/internal/pop3"
	"github.com/wispmail/wispmail/internal/queue"
	"github.com/wispmail/wispmail/internal/sieve"
	"github.com/wispmail/wispmail/internal/smtp"
	"github.com/wispmail/wispmail/internal/spf"
	"github.com/wispmail/wispmail/internal/storage/elasticsearch"
	"github.com/wispmail/wispmail/internal/storage/postgres"
	"github.com/wispmail/wispmail/internal/storage/redis"
	"github.com/wispmail/wispmail/internal/storage/s3"
	"github.com/wispmail/wispmail/pkg/email"
	"github.com/wispmail/wispmail/pkg/graceful"
	"github.com/wispmail/wispmail/pkg/logger"
	"github.com/wispmail/wispmail/pkg/metrics"
	"github.com/wispmail/wispmail/pkg/tracing"
)

type App struct {
	cfg         *config.Config
	logger      *logger.Logger
	metrics     *metrics.Metrics
	tracer      *tracing.Provider
	db          *postgres.Client
	redisClient *redis.Client
	s3Client    *s3.Client
	esClient    *elasticsearch.Client
	keyManager  *jwt.KeyManager
	tokenSvc    *jwt.TokenService
	enforcer    *rbac.Enforcer
	apiKeyVal   *apikey.Validator
	dkimSigner  *dkim.Signer
	spfVal      *spf.Validator
	dmarcVal    *dmarc.Validator
	emailVal    *email.Validator
	apiServer   *api.Server
	smtpServer  *smtp.Server
	imapServer  *imap.Server
	pop3Server  *pop3.Server
	jmapServer  *jmap.Server
	sieveEng    *sieve.Engine
	queueSrv    *queue.Server
	worker      *queue.Worker
	shutdown    *graceful.Manager
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	app := &App{
		cfg:      cfg,
		shutdown: graceful.NewManager(cfg.API.ShutdownTimeout),
	}

	if err := app.initLogger(); err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}
	if err := app.initMetrics(); err != nil {
		return nil, fmt.Errorf("metrics: %w", err)
	}
	if err := app.initTracer(ctx); err != nil {
		return nil, fmt.Errorf("tracer: %w", err)
	}
	if err := app.initDatabase(ctx); err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}
	if err := app.initRedis(ctx); err != nil {
		return nil, fmt.Errorf("redis: %w", err)
	}
	if err := app.initS3(ctx); err != nil {
		return nil, fmt.Errorf("s3: %w", err)
	}
	app.initElasticsearch()
	app.initAuth()
	app.initSecurity()
	app.initEmailTools()
	if err := app.initQueue(ctx); err != nil {
		return nil, fmt.Errorf("queue: %w", err)
	}
	app.initAPIServer()
	app.initSMTPServer()
	app.initIMAPServer()
	app.initPOP3Server()
	app.initJMAPServer()
	app.initSieveEngine()
	app.registerShutdownHooks()

	return app, nil
}

func (a *App) initLogger() error {
	a.logger = logger.New(logger.Config{
		Level:  a.cfg.LogLevel,
		Format: a.cfg.LogFormat,
	})
	a.logger.Info("logger initialized", map[string]any{"level": a.cfg.LogLevel})
	return nil
}

func (a *App) initMetrics() error {
	a.metrics = metrics.New()
	a.metrics.RegisterCounter("wispmail_emails_sent_total", "Total emails sent", []string{"tenant_id", "status"})
	a.metrics.RegisterGauge("wispmail_active_connections", "Active connections", []string{"protocol"})
	a.metrics.RegisterHistogram("wispmail_request_duration_seconds", "Request duration", []string{"method", "path"}, []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10})
	return nil
}

func (a *App) initTracer(ctx context.Context) error {
	tracer, err := tracing.NewProvider(ctx, tracing.Config{
		ExporterEndpoint: a.cfg.OTEL.ExporterEndpoint,
		ServiceName:      a.cfg.OTEL.ServiceName,
		ServiceVersion:   a.cfg.OTEL.ServiceVersion,
		Enabled:          a.cfg.OTEL.ExporterEndpoint != "",
		SampleRate:       1.0,
		BatchTimeout:     5 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("failed to create tracer: %w", err)
	}
	a.tracer = tracer
	return nil
}

func (a *App) initDatabase(ctx context.Context) error {
	client, err := postgres.NewClient(ctx, a.cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	a.db = client
	a.logger.Info("database connected", nil)
	return nil
}

func (a *App) initRedis(ctx context.Context) error {
	client, err := redis.NewClient(ctx, a.cfg.Redis.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	a.redisClient = client
	a.logger.Info("redis connected", nil)
	return nil
}

func (a *App) initS3(ctx context.Context) error {
	client, err := s3.NewClient(ctx, s3.S3Config{
		Endpoint:        a.cfg.S3.Endpoint,
		Region:          a.cfg.S3.Region,
		BucketName:      a.cfg.S3.BucketName,
		AccessKeyID:     a.cfg.S3.AccessKeyID,
		SecretAccessKey: a.cfg.S3.SecretAccessKey,
		UseSSL:          a.cfg.S3.UseSSL,
		ForcePathStyle:  a.cfg.S3.ForcePathStyle,
	})
	if err != nil {
		return fmt.Errorf("failed to create S3 client: %w", err)
	}
	a.s3Client = client
	a.logger.Info("s3 client initialized", nil)
	return nil
}

func (a *App) initElasticsearch() {
	a.esClient = elasticsearch.NewClient(elasticsearch.Config{
		URL:         a.cfg.Elasticsearch.URL,
		IndexPrefix: a.cfg.Elasticsearch.IndexPrefix,
	})
}

func (a *App) initAuth() {
	a.keyManager = jwt.NewKeyManager()
	if _, err := a.keyManager.GenerateKeyPair(); err != nil {
		a.logger.Error("failed to generate JWT key pair", err, nil)
		os.Exit(1)
	}
	a.tokenSvc = jwt.NewTokenService(a.keyManager, a.cfg.JWT.Issuer, a.cfg.JWT.AccessTTL, a.cfg.JWT.RefreshTTL)
	a.enforcer = rbac.NewEnforcer()
	a.apiKeyVal = apikey.NewValidator(apikey.NewHasher(), nil)
	a.logger.Info("authentication initialized", nil)
}

func (a *App) initSecurity() {
	var err error
	a.dkimSigner, err = dkim.NewSigner(dkim.DKIMConfig{
		Selector:       a.cfg.DKIM.Selector,
		Domain:         a.cfg.DKIM.Domain,
		PrivateKeyPath: a.cfg.DKIM.PrivateKeyPath,
	})
	if err != nil {
		a.logger.Warn("DKIM signer initialization failed", map[string]any{"error": err.Error()})
	}
	dnsResolver := spf.NewDefaultDNSResolver()
	a.spfVal = spf.NewValidator(dnsResolver)
	dmarcResolver := dmarc.NewDefaultDNSResolver()
	a.dmarcVal = dmarc.NewValidator(dmarcResolver)
	a.logger.Info("email security initialized", nil)
}

func (a *App) initEmailTools() {
	a.emailVal = email.NewValidator()
}

func (a *App) initQueue(ctx context.Context) error {
	redisStore := redis.NewQueueStore(a.redisClient)
	a.queueSrv, _ = queue.NewServer(redisStore, queue.ServerConfig{
		Concurrency: 10,
		Queues:      map[string]int{"critical": 6, "default": 3, "low": 1},
	})
	a.worker = queue.NewWorker(redisStore, a.db.Pool, queue.WorkerConfig{
		Concurrency: 10,
		Queues:      map[string]int{"critical": 6, "default": 3, "low": 1},
	})
	a.logger.Info("queue system initialized", nil)
	return nil
}

func (a *App) initAPIServer() {
	a.apiServer = api.NewServer(api.ServerConfig{
		Host:         a.cfg.API.Host,
		Port:         a.cfg.API.Port,
		ReadTimeout:  a.cfg.API.ReadTimeout,
		WriteTimeout: a.cfg.API.WriteTimeout,
		Metrics:      a.metrics,
		Logger:       a.logger,
		TokenService: a.tokenSvc,
		Enforcer:     a.enforcer,
	})
}

func (a *App) initSMTPServer() {
	a.smtpServer = smtp.NewServer(smtp.Config{
		Host:           a.cfg.SMTP.Host,
		Port:           a.cfg.SMTP.Port,
		TLSCert:        a.cfg.SMTP.TLSCert,
		TLSKey:         a.cfg.SMTP.TLSKey,
		MaxMessageSize: a.cfg.SMTP.MaxMessageSize,
		Logger:         a.logger,
		Metrics:        a.metrics,
	})
}

func (a *App) initIMAPServer() {
	a.imapServer = imap.NewServer(imap.Config{
		Host:    a.cfg.IMAP.Host,
		Port:    a.cfg.IMAP.Port,
		TLSCert: a.cfg.IMAP.TLSCert,
		TLSKey:  a.cfg.IMAP.TLSKey,
		Logger:  a.logger,
	})
}

func (a *App) initPOP3Server() {
	a.pop3Server = pop3.NewServer(pop3.Config{
		Host:    a.cfg.POP3.Host,
		Port:    a.cfg.POP3.Port,
		TLSCert: a.cfg.POP3.TLSCert,
		TLSKey:  a.cfg.POP3.TLSKey,
	})
}

func (a *App) initJMAPServer() {
	a.jmapServer = jmap.NewServer(jmap.Config{
		Host:    a.cfg.JMAP.Host,
		Port:    a.cfg.JMAP.Port,
		TLSCert: a.cfg.JMAP.TLSCert,
		TLSKey:  a.cfg.JMAP.TLSKey,
	})
}

func (a *App) initSieveEngine() {
	a.sieveEng = sieve.NewEngine()
}

func (a *App) registerShutdownHooks() {
	a.shutdown.Add(func(ctx context.Context) error { return a.apiServer.Shutdown(ctx) })
	a.shutdown.Add(func(ctx context.Context) error { return a.smtpServer.Shutdown(ctx) })
	if a.imapServer != nil {
		a.shutdown.Add(func(ctx context.Context) error { return a.imapServer.Shutdown(ctx) })
	}
	if a.pop3Server != nil {
		a.shutdown.Add(func(ctx context.Context) error { return a.pop3Server.Shutdown(ctx) })
	}
	if a.jmapServer != nil {
		a.shutdown.Add(func(ctx context.Context) error { return a.jmapServer.Shutdown(ctx) })
	}
	a.shutdown.Add(func(ctx context.Context) error { return a.worker.Stop(ctx) })
	a.shutdown.Add(func(ctx context.Context) error { return a.queueSrv.Shutdown(ctx) })
	a.shutdown.Add(func(ctx context.Context) error { a.db.Close(); return nil })
	a.shutdown.Add(func(ctx context.Context) error { return a.redisClient.Close() })
	a.shutdown.Add(func(ctx context.Context) error { return a.tracer.Shutdown(ctx) })
}

func (a *App) Start(ctx context.Context) error {
	a.logger.Info("starting Wispmail server", map[string]any{"env": a.cfg.Env})
	go func() { a.worker.Start(ctx) }()
	go func() { a.queueSrv.Start(ctx) }()
	go func() { a.smtpServer.Start() }()
	if a.imapServer != nil {
		go func() { a.imapServer.Start() }()
	}
	if a.pop3Server != nil {
		go func() { a.pop3Server.Start() }()
	}
	if a.jmapServer != nil {
		go func() { a.jmapServer.Start() }()
	}
	return a.apiServer.Start()
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("stopping Wispmail server", nil)
	a.shutdown.Shutdown()
	return nil
}