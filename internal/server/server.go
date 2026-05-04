package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/ajawes/hesp/internal/ingestion/handler"
)

type Server struct {
	httpServer *http.Server
}

func New() *Server {
	mux := http.NewServeMux()

	handler := handler.NewHandler()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("/events/ingest", handler)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: mux,
		},
	}
}

func (s *Server) Start() error {
	log.Printf("starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Printf("stopping server")
	return s.httpServer.Shutdown(ctx)
}
