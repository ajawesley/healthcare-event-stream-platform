package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ajawes/hesp/internal/ingestion/server"
	"github.com/ajawes/hesp/internal/observability"
)

func main() {
	// -----------------------------
	// Observability
	// -----------------------------
	logger := observability.NewLogger("hesp-ecs", "dev")
	defer logger.Sync()

	observability.InitMetrics("hesp-ecs", "dev")

	// ⭐ Initialize tracing BEFORE creating the server
	shutdownTracing := observability.InitTracing("hesp-ecs", "v1.0.0", "dev")
	defer shutdownTracing(context.Background())

	// Now the server, router, handlers, dispatcher all see a real tracer
	srv := server.New()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-stop
	if err := srv.Stop(); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
}
