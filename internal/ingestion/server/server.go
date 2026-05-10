package server

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/compliance"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/handler"
	"github.com/ajawes/hesp/internal/ingestion/router"
	"github.com/ajawes/hesp/internal/observability"
)

type Server struct {
	httpServer      *http.Server
	router          *router.FormatRouter
	complianceGuard *compliance.Guard
	complianceDB    compliance.ClientAPI
	shutdownFn      func(context.Context) error
}

type Option func(*Server)

func WithComplianceGuard(g *compliance.Guard) Option {
	return func(s *Server) { s.complianceGuard = g }
}

func WithComplianceClient(c compliance.ClientAPI) Option {
	return func(s *Server) { s.complianceDB = c }
}

func New(opts ...Option) *Server {
	s := &Server{}

	for _, opt := range opts {
		opt(s)
	}

	mux := http.NewServeMux()

	// Build base ingestion router (guard injected later if needed)
	ingestRouter := router.NewFormatRouter(
		router.WithDetector(detector.NewDetector()),
		router.WithNormalizationRouter(router.NewNormalizationRouter()),
		router.WithTransformationRouter(router.NewTransformationRouter()),
	)

	ingestHandler := handler.NewHandler(
		handler.WithRouter(ingestRouter),
	)

	// -------------------------------------------------------------------------
	// Health endpoints
	// -------------------------------------------------------------------------

	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if s.complianceDB != nil {
			if err := s.complianceDB.Ready(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("compliance db not ready"))
				return
			}
		}

		if s.router == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("router not ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("/events/ingest", ingestHandler)

	s.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	s.router = ingestRouter

	return s
}

func (s *Server) Start() error {
	// -----------------------------
	// Observability
	// -----------------------------
	observability.NewLogger("hesp-ecs", "dev")
	observability.InitMetrics("hesp-ecs", "dev")

	shutdownTracing := observability.InitTracing(
		"hesp-ecs",
		"v1.0.0",
		"dev",
	)
	s.shutdownFn = shutdownTracing

	// -----------------------------
	// Compliance DB Wiring
	// -----------------------------
	if s.complianceDB == nil {
		rawPassword := config.GetEnv("COMPLIANCE_DB_PASSWORD", "")
		escapedPassword := url.QueryEscape(rawPassword)

		dbCfg := compliance.Config{
			Host:     config.GetEnv("COMPLIANCE_DB_HOST", ""),
			Port:     config.GetEnvInt("COMPLIANCE_DB_PORT", 5432),
			User:     config.GetEnv("COMPLIANCE_DB_USER", ""),
			Password: escapedPassword,
			Database: config.GetEnv("COMPLIANCE_DB_NAME", ""),
		}

		compClient, err := compliance.NewClient(context.Background(), dbCfg)
		if err != nil {
			log.Fatalf("failed to initialize compliance db client: %v", err)
		}

		s.complianceDB = compClient
	}

	// -----------------------------
	// Compliance Guard Wiring
	// -----------------------------
	if s.complianceGuard == nil {
		breaker := compliance.NewCircuitBreaker(5, 30*time.Second)
		s.complianceGuard = compliance.NewGuard(s.complianceDB, breaker)
	}

	// Rebuild router WITH guard
	s.router = router.NewFormatRouter(
		router.WithDetector(detector.NewDetector()),
		router.WithNormalizationRouter(router.NewNormalizationRouter()),
		router.WithTransformationRouter(router.NewTransformationRouter()),
		router.WithComplianceGuard(s.complianceGuard),
	)

	// -----------------------------
	// AWS + Dispatcher Wiring
	// -----------------------------
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	s3cfg := config.LoadS3Config()

	s.router.SetDispatcher(
		dispatcher.NewS3Dispatcher(
			s3Client,
			s3cfg.Bucket,
			s3cfg.Prefix,
			s3cfg.KMSKeyARN,
		),
	)

	// -----------------------------
	// Observability Middleware
	// -----------------------------
	s.httpServer.Handler = observability.ObservabilityMiddleware(s.httpServer.Handler)

	log.Printf("starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("stopping server")

	if s.shutdownFn != nil {
		_ = s.shutdownFn(ctx)
	}

	return s.httpServer.Shutdown(ctx)
}
