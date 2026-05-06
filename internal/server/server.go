package server

import (
	"context"
	"log"
	"net/http"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/handler"
	"github.com/ajawes/hesp/internal/ingestion/router"
	"github.com/ajawes/hesp/internal/observability"
)

type Server struct {
	httpServer *http.Server
	router     *router.FormatRouter
	shutdownFn func(context.Context) error
}

func New() *Server {
	mux := http.NewServeMux()

	// Build ingestion router WITHOUT dispatcher (we inject it later in Start)
	ingestRouter := router.NewFormatRouter(
		router.WithDetector(detector.NewDetector()),
		router.WithNormalizationRouter(router.NewNormalizationRouter()),
		router.WithTransformationRouter(router.NewTransformationRouter()),
	)

	ingestHandler := handler.NewHandler(
		handler.WithRouter(ingestRouter),
	)

	// Health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Ingestion endpoint
	mux.Handle("/events/ingest", ingestHandler)

	return &Server{
		httpServer: &http.Server{
			Addr: ":8080",
			// Middleware will be injected in Start()
			Handler: mux,
		},
		router: ingestRouter,
	}
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
	// (This will be replaced with your new ECS middleware)
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
