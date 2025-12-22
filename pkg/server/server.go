package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	cfg := &Config{
		Address:         "",
		Port:            8080,
		RateLimit:       100, // 100 req/s
		RateLimitBurst:  200, // burst of 200
		CacheMaxAge:     300, // 5 minutes
		MaxBulkRequests: 100,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		LogLevel:        slog.LevelInfo,
	}

	// Override with environment variables if set
	if portStr := os.Getenv("PORT"); portStr != "" {
		var port int
		if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
			cfg.Port = port
		}
	}

	if logLevelStr := os.Getenv("LOG_LEVEL"); logLevelStr != "" {
		var level slog.Level
		if err := level.UnmarshalText([]byte(logLevelStr)); err == nil {
			cfg.LogLevel = level
		}
	}

	return cfg
}

// Server represents the HTTP server
type Server struct {
	config      *Config
	httpServer  *http.Server
	rateLimiter *rate.Limiter
	mu          sync.RWMutex
	ready       bool
	logger      Logger
	validator   *Validator
}

// NewServer creates a new server instance
func NewServer(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	s := &Server{
		config:      config,
		rateLimiter: rate.NewLimiter(config.RateLimit, config.RateLimitBurst),
		logger:      NewLogger(slog.LevelInfo),
		validator:   NewValidator(),
	}

	// Setup HTTP server
	mux := s.setupRoutes()
	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Address, config.Port),
		Handler:      mux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
	}

	return s
}

// setupRoutes configures all HTTP routes and middleware
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// System endpoints (no rate limiting)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/ready", s.handleReady)

	// API endpoints with middleware
	mux.HandleFunc("/v1/recommendations", s.withMiddleware(s.handleGetRecommendations))
	mux.HandleFunc("/v1/recommendations/resolve", s.withMiddleware(s.handleBulkResolve))

	return mux
}

// writeJSON writes JSON response
func (s *Server) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode JSON response: %v\n", err)
	}
}

// writeError writes error response
func (s *Server) writeError(w http.ResponseWriter, r *http.Request, statusCode int,
	code, message string, retryable bool, details map[string]interface{}) {

	requestID, _ := r.Context().Value(contextKeyRequestID).(string)
	if requestID == "" {
		requestID = uuid.New().String()
	}

	errResp := ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
		Retryable: retryable,
	}

	s.writeJSON(w, statusCode, errResp)
}

// SetReady marks the server as ready to serve traffic
func (s *Server) SetReady(ready bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ready = ready
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	s.SetReady(true)

	fmt.Printf("Starting server on %s\n", s.httpServer.Addr)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		return s.Shutdown(context.Background())
	case err := <-errChan:
		return err
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.SetReady(false)

	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	fmt.Println("Shutting down server...")
	return s.httpServer.Shutdown(shutdownCtx)
}

// Run starts the server with graceful shutdown handling
func Run() error {
	return RunWithConfig(DefaultConfig())
}

// RunWithConfig starts the server with custom configuration
func RunWithConfig(config *Config) error {
	server := NewServer(config)

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Use errgroup for concurrent operations
	g, gctx := errgroup.WithContext(ctx)

	// Start HTTP server
	g.Go(func() error {
		return server.Start(gctx)
	})

	// Wait for completion or error
	if err := g.Wait(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	fmt.Println("Server stopped gracefully")
	return nil
}

// Utility functions

const defaultQueryValue = "ALL"

func getQueryParamOrDefault(q map[string][]string, key string) string {
	if values, ok := q[key]; ok && len(values) > 0 && values[0] != "" {
		return values[0]
	}
	return defaultQueryValue
}

func stringPtr(s string) *string {
	return &s
}
