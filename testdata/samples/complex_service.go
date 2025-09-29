package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
)

type Config struct {
	Port     string
	Database string
}

func NewServer(cfg Config) *Server {
	return &Server{config: cfg}
}

type Server struct {
	config Config
}

func (s *Server) Start() error {
	// Dynamic port from config
	addr := fmt.Sprintf(":%s", s.config.Port)
	return http.ListenAndServe(addr, s.routes())
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Health check makes external call
		resp, err := http.Get("http://health-service.internal:8080/status")
		if err != nil {
			w.WriteHeader(500)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(200)
	})

	return mux
}

func (s *Server) connectDatabase() error {
	// Database connection using config
	conn, err := net.Dial("tcp", s.config.Database)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}

func startMetricsServer() {
	// Metrics server on environment variable port
	port := os.Getenv("METRICS_PORT")
	if port == "" {
		port = "9090"
	}
	
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return
	}
	defer listener.Close()
}

func backgroundWorker(ctx context.Context) {
	// Worker makes API calls
	client := &http.Client{}
	req, _ := http.NewRequestWithContext(ctx, "POST", "https://api.external.com/webhook", nil)
	client.Do(req)
}