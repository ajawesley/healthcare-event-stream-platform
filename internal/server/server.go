package server

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/handler"
	"github.com/ajawes/hesp/internal/ingestion/router"
)

type Server struct {
	httpServer *http.Server
	router     *router.FormatRouter
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

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("/events/ingest", ingestHandler)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
		router: ingestRouter,
	}
}

func (s *Server) Start() error {
	logger := slog.Default()

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)

	// Load S3 config *here*, not in New()
	s3cfg := config.LoadS3Config()

	// Inject dispatcher now that config is available
	s.router.SetDispatcher(
		dispatcher.NewS3Dispatcher(
			s3Client,
			s3cfg.Bucket,
			s3cfg.Prefix,
			s3cfg.KMSKeyARN,
			logger,
		),
	)

	log.Printf("starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Printf("stopping server")
	return s.httpServer.Shutdown(ctx)
}
