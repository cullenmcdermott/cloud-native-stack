// Package api provides the HTTP API layer for the CNS System Configuration Recommendation service.
//
// This package acts as a thin wrapper around the reusable pkg/server package,
// configuring it with application-specific routes and handlers.
//
// # Usage
//
// To start the API server:
//
//	package main
//
//	import (
//	    "log"
//	    "github.com/NVIDIA/cloud-native-stack/pkg/api"
//	)
//
//	func main() {
//	    if err := api.Serve(); err != nil {
//	        log.Fatalf("server error: %v", err)
//	    }
//	}
//
// # Architecture
//
// The API layer is responsible for:
//   - Configuring structured logging with application name and version
//   - Setting up route handlers (e.g., /v1/recommendations)
//   - Delegating server lifecycle management to pkg/server
//
// The pkg/server package handles:
//   - HTTP server setup and graceful shutdown
//   - Middleware (rate limiting, logging, metrics, panic recovery)
//   - Health and readiness endpoints
//   - Prometheus metrics
//
// # Endpoints
//
// Application Endpoints (with rate limiting):
//   - GET /v1/recommendations - Get system configuration recommendations
//
// System Endpoints (no rate limiting):
//   - GET /health  - Health check
//   - GET /ready   - Readiness check
//   - GET /metrics - Prometheus metrics
//
// # Configuration
//
// The server is configured via environment variables:
//   - PORT: HTTP server port (default: 8080)
//
// Version information is set at build time using ldflags:
//
//	go build -ldflags="-X 'github.com/NVIDIA/cloud-native-stack/pkg/api.version=1.0.0'"
package api
