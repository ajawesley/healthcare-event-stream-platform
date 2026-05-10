package server

import (
	"context"
	"log"
	"net/http"
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

	// Build ingestion router WITHOUT dispatcher (we inject it later in Start)
	ingestRouter := router.NewFormatRouter(
		router.WithDetector(detector.NewDetector()),
		router.WithNormalizationRouter(router.NewNormalizationRouter()),
		router.WithTransformationRouter(router.NewTransformationRouter()),
	)

	// Inject compliance guard if provided
	if s.complianceGuard != nil {
		ingestRouter = router.NewFormatRouter(
			router.WithDetector(detector.NewDetector()),
			router.WithNormalizationRouter(router.NewNormalizationRouter()),
			router.WithTransformationRouter(router.NewTransformationRouter()),
			router.WithComplianceGuard(s.complianceGuard),
		)
	}

	ingestHandler := handler.NewHandler(
		handler.WithRouter(ingestRouter),
	)

	// -------------------------------------------------------------------------
	// Health endpoints
	// -------------------------------------------------------------------------

	// Liveness — container is running
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})

	// Readiness — dependencies ready
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check compliance DB readiness
		if s.complianceDB != nil {
			if err := s.complianceDB.Ready(r.Context()); err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("compliance db not ready"))
				return
			}
		}

		// Check router + dispatcher
		if s.router == nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("router not ready"))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	// Existing health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Ingestion endpoint
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
	// Initialize Observability
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
	dbCfg := compliance.Config{
		Host:     config.GetEnv("COMPLIANCE_DB_HOST", ""),
		Port:     config.GetEnvInt("COMPLIANCE_DB_PORT", 5432),
		User:     config.GetEnv("COMPLIANCE_DB_USER", ""),
		Password: config.GetEnv("COMPLIANCE_DB_PASSWORD", ""),
		Database: config.GetEnv("COMPLIANCE_DB_NAME", ""),
	}

	compClient, err := compliance.NewClient(context.Background(), dbCfg)
	if err != nil {
		log.Fatalf("failed to initialize compliance db client: %v", err)
	}

	s.complianceDB = compClient

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
	// Inject Observability Middleware
	// -----------------------------
	s.httpServer.Handler = observability.ObservabilityMiddleware(s.httpServer.Handler)

	log.Printf("starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("stopping server")

	// Shutdown tracing provider
	if s.shutdownFn != nil {
		_ = s.shutdownFn(ctx)
	}

	return s.httpServer.Shutdown(ctx)
}
