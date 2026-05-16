package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ajawes/hesp/internal/ingestion/server"
	"github.com/ajawes/hesp/internal/observability"
)

func main() {
	// -----------------------------
	// Observability (still works locally)
	// -----------------------------
	logger := observability.NewLogger("hesp-ecs", "dev")
	defer logger.Sync()

	observability.InitMetrics("hesp-ecs", "dev")

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
