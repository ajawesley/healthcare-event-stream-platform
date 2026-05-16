package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"

	"github.com/ajawes/hesp/internal/config"
	"github.com/ajawes/hesp/internal/ingestion/compliance"
	"github.com/ajawes/hesp/internal/ingestion/detector"
	"github.com/ajawes/hesp/internal/ingestion/dispatcher"
	"github.com/ajawes/hesp/internal/ingestion/handler"
	"github.com/ajawes/hesp/internal/ingestion/router"
	"github.com/ajawes/hesp/internal/observability"
)

// -----------------------------------------------------------------------------
// Readiness Framework
// -----------------------------------------------------------------------------

type ReadyCheck interface {
	Ready(ctx context.Context) bool
	Name() string
}

type readyFunc struct {
	name string
	fn   func(ctx context.Context) bool
}

func (r readyFunc) Ready(ctx context.Context) bool { return r.fn(ctx) }
func (r readyFunc) Name() string                   { return r.name }

// -----------------------------------------------------------------------------
// Server
// -----------------------------------------------------------------------------

type Server struct {
	httpServer      *http.Server
	router          router.Router
	complianceGuard compliance.ComplianceGuard
	complianceDB    compliance.ClientAPI
	shutdownFn      func(context.Context) error

	readyChecks []ReadyCheck
}

type Option func(*Server)

// global tracer for this package
var tracer = otel.Tracer("hesp-ecs/server")

func WithComplianceGuard(g compliance.ComplianceGuard) Option {
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

	// Base ingestion router (guard injected later)
	s.router = router.NewFormatRouter()

	ingestHandler := handler.NewHandler(
		handler.WithRouter(s.router),
	)
	mux.Handle("/events/ingest", ingestHandler)

	// -------------------------------------------------------------------------
	// Health Endpoints
	// -------------------------------------------------------------------------

	// Liveness now checks panic state
	mux.HandleFunc("/livez", func(w http.ResponseWriter, r *http.Request) {
		if panicHandler.panic != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("panic detected"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("alive"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		for _, chk := range s.readyChecks {
			if !chk.Ready(ctx) {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte(chk.Name() + " not ready"))
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ready"))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	s.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	// -----------------------------
	// Detect Local Mode
	// -----------------------------
	isLocal := config.GetEnv("LOCAL_MODE", "") == "true"

	shutdownTracing := observability.InitTracing(
		"hesp-ecs",
		"v1.0.0",
		"dev",
	)
	s.shutdownFn = shutdownTracing

	// -----------------------------
	// Readiness: Lineage
	// -----------------------------
	if !isLocal {
		s.readyChecks = append(s.readyChecks, readyFunc{
			name: "lineage",
			fn: func(ctx context.Context) bool {
				ctx, span := tracer.Start(ctx, "readiness.lineage")
				defer span.End()
				observability.ObserveLineageLatency(ctx, "readiness_probe", time.Now())
				return true
			},
		})
	}

	// -----------------------------
	// Panic readiness
	// -----------------------------
	defer func() {
		if r := recover(); r != nil {
			log.Printf("temporarily recovering from panic %+v", r)
			PanicHandler.Set(r)
		}
	}()

	if !isLocal {
		s.readyChecks = append(s.readyChecks, readyFunc{
			name: "catastrophe",
			fn: func(ctx context.Context) bool {
				return PanicHandler.Get() == nil
			},
		})
	}

	// -----------------------------
	// Build Compliance Subsystem
	// -----------------------------
	if isLocal {
		// -----------------------------------------
		// LOCAL MODE: No AWS, No DB, No Redis
		// -----------------------------------------
		fmt.Println("⚠️  LOCAL MODE ENABLED — using mock compliance DB + guard")

		s.complianceDB = compliance.NewMockClient() // you already have this pattern
		s.complianceGuard = compliance.NewNoopGuard()

	} else {
		// -----------------------------------------
		// NORMAL MODE (AWS + DB)
		// -----------------------------------------
		if s.complianceDB == nil {
			rawPassword := config.GetEnv("COMPLIANCE_DB_PASSWORD", "")
			escapedPassword := url.QueryEscape(rawPassword)

			pgURL := fmt.Sprintf(
				"postgres://%s:%s@%s:%d/%s",
				config.GetEnv("COMPLIANCE_DB_USER", ""),
				escapedPassword,
				config.GetEnv("COMPLIANCE_DB_HOST", ""),
				config.GetEnvInt("COMPLIANCE_DB_PORT", 5432),
				config.GetEnv("COMPLIANCE_DB_NAME", ""),
			)

			log.Printf("connecting to postgres at %s", pgURL)

			pool, err := pgxpool.New(context.Background(), pgURL)
			if err != nil {
				log.Fatalf("failed to initialize postgres pool: %v", err)
			}

			log.Printf("created postgres pool %+v", pool)

			awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
			if err != nil {
				log.Fatalf("failed to load AWS config: %v", err)
			}
			dynClient := dynamodb.NewFromConfig(awsCfg)

			ttl := time.Duration(config.GetEnvInt("REDIS_TTL_SECONDS", 300)) * time.Second

			pgStore := compliance.NewPostgresStore(pool)
			dynStore := compliance.NewDynamoStore(dynClient, config.GetEnv("DYNAMO_TABLE", "compliance_rules"))
			redisStore := compliance.NewRedisStore(config.GetEnv("REDIS_ADDR", "localhost:6379"), ttl)

			s.complianceDB = compliance.NewClient(pgStore, dynStore, redisStore)

			log.Printf("created compliance DB %+v", s.complianceDB)

			// Readiness checks
			s.readyChecks = append(s.readyChecks, readyFunc{
				name: "postgres",
				fn: func(ctx context.Context) bool {
					ctx, span := tracer.Start(ctx, "readiness.postgres")
					defer span.End()
					return pool.Ping(ctx) == nil
				},
			})

			s.readyChecks = append(s.readyChecks, readyFunc{
				name: "dynamodb",
				fn: func(ctx context.Context) bool {
					ctx, span := tracer.Start(ctx, "readiness.dynamodb")
					defer span.End()
					limit := int32(1)
					_, err := dynClient.ListTables(ctx, &dynamodb.ListTablesInput{Limit: &limit})
					return err == nil
				},
			})

			s.readyChecks = append(s.readyChecks, readyFunc{
				name: "redis",
				fn: func(ctx context.Context) bool {
					ctx, span := tracer.Start(ctx, "readiness.redis")
					defer span.End()
					return redisStore.Ping(ctx) == nil
				},
			})
		}

		if s.complianceGuard == nil {
			s.complianceGuard = compliance.NewGuard(s.complianceDB)
		}
	}

	// -----------------------------
	// Set Up Router
	// -----------------------------
	r, ok := (*router.FormatRouter)(nil), false
	if r, ok = s.router.(*router.FormatRouter); !ok {
		log.Fatal("router is not a FormatRouter")
	}
	r.SetDetector(detector.NewDetector())
	r.SetNormalizer(router.NewNormalizationRouter())
	r.SetTransformer(router.NewTransformationRouter())
	r.SetComplianceGuard(s.complianceGuard)

	// -----------------------------
	// Dispatcher
	// -----------------------------
	if isLocal {
		fmt.Println("⚠️  LOCAL MODE — using NoopDispatcher (no S3 calls)")
		r.SetDispatcher(dispatcher.NewNoopDispatcher())
	} else {
		awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
		if err != nil {
			log.Fatalf("failed to load AWS config: %v", err)
		}
		s3Client := s3.NewFromConfig(awsCfg)
		s3cfg := config.LoadS3Config()

		r.SetDispatcher(
			dispatcher.NewS3Dispatcher(
				s3Client,
				s3cfg.Bucket,
				s3cfg.Prefix,
				s3cfg.KMSKeyARN,
			),
		)

		s.readyChecks = append(s.readyChecks, readyFunc{
			name: "s3",
			fn: func(ctx context.Context) bool {
				ctx, span := tracer.Start(ctx, "readiness.s3")
				defer span.End()
				_, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
				return err == nil
			},
		})
	}

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
