package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	routes := map[string]http.HandlerFunc{
		"/test": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	s := New(routes)
	if s == nil {
		t.Fatal("expected server instance, got nil")
	}

	if s.config == nil {
		t.Error("expected config to be initialized")
	}

	if s.httpServer == nil {
		t.Error("expected httpServer to be initialized")
	}

	if s.rateLimiter == nil {
		t.Error("expected rateLimiter to be initialized")
	}
}

func TestHealthEndpoint(t *testing.T) {
	routes := map[string]http.HandlerFunc{}
	s := New(routes)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
	}
}

func TestReadyEndpoint(t *testing.T) {
	routes := map[string]http.HandlerFunc{}
	s := New(routes)

	tests := []struct {
		name           string
		ready          bool
		expectedStatus int
	}{
		{
			name:           "ready state",
			ready:          true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not ready state",
			ready:          false,
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.setReady(tt.ready)

			req := httptest.NewRequest(http.MethodGet, "/ready", nil)
			w := httptest.NewRecorder()

			s.handleReady(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestRateLimiting(t *testing.T) {
	routes := map[string]http.HandlerFunc{
		"/test": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	// Create a custom config with very restrictive rate limiting
	cfg := NewConfig()
	cfg.RateLimit = 1      // 1 req/sec
	cfg.RateLimitBurst = 1 // burst of 1

	s := New(routes, WithConfig(cfg))

	handler := s.withMiddleware(routes["/test"])

	// First request should succeed
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w1 := httptest.NewRecorder()
	handler(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("expected first request to succeed with status 200, got %d", w1.Code)
	}

	// Second request should be rate limited (bucket is empty)
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	handler(w2, req2)

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected rate limit error with status 429, got %d", w2.Code)
	}

	if w2.Header().Get("Retry-After") == "" {
		t.Error("expected Retry-After header to be set")
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	routes := map[string]http.HandlerFunc{
		"/test": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	s := New(routes)

	t.Run("generates request ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler := s.requestIDMiddleware(routes["/test"])
		handler(w, req)

		requestID := w.Header().Get("X-Request-Id")
		if requestID == "" {
			t.Error("expected X-Request-Id header to be set")
		}
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		expectedID := "550e8400-e29b-41d4-a716-446655440000"
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-Id", expectedID)
		w := httptest.NewRecorder()

		handler := s.requestIDMiddleware(routes["/test"])
		handler(w, req)

		requestID := w.Header().Get("X-Request-Id")
		if requestID != expectedID {
			t.Errorf("expected request ID %s, got %s", expectedID, requestID)
		}
	})

	t.Run("regenerates invalid UUID", func(t *testing.T) {
		invalidID := "not-a-valid-uuid"
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-Id", invalidID)
		w := httptest.NewRecorder()

		handler := s.requestIDMiddleware(routes["/test"])
		handler(w, req)

		requestID := w.Header().Get("X-Request-Id")
		if requestID == invalidID {
			t.Error("expected invalid UUID to be regenerated")
		}
	})
}

func TestPanicRecovery(t *testing.T) {
	panicHandler := func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	}

	routes := map[string]http.HandlerFunc{
		"/panic": panicHandler,
	}

	s := New(routes)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	handler := s.panicRecoveryMiddleware(panicHandler)

	// Should not panic, should return 500
	handler(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d after panic recovery, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGracefulShutdown(t *testing.T) {
	routes := map[string]http.HandlerFunc{
		"/test": func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	}

	s := New(routes)
	s.config.ShutdownTimeout = 100 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Start(ctx)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	// Cancel context to trigger shutdown
	cancel()

	// Wait for shutdown to complete
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("expected clean shutdown, got error: %v", err)
		}
	case <-time.After(time.Second):
		t.Error("shutdown timed out")
	}
}
