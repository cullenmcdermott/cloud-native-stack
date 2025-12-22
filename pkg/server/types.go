package server

import (
	"log/slog"
	"time"

	"golang.org/x/time/rate"
)

// Domain types matching OpenAPI spec schemas

// RecommendationRequest represents the input parameters for recommendation queries
type RecommendationRequest struct {
	OSFamily                string  `json:"osFamily"`
	OSVersion               string  `json:"osVersion"`
	Kernel                  string  `json:"kernel"`
	Environment             string  `json:"environment"`
	Kubernetes              string  `json:"kubernetes"`
	GPU                     string  `json:"gpu"`
	Intent                  string  `json:"intent"`
	PayloadVersionRequested *string `json:"payloadVersionRequested"`
}

// ComponentRecommendation represents a single component with version
type ComponentRecommendation struct {
	Name    string  `json:"name"`
	Version *string `json:"version"`
	Notes   *string `json:"notes,omitempty"`
}

// CNSReleaseRecommendation represents a CNS release with components
type CNSReleaseRecommendation struct {
	CNSVersion  string                    `json:"cnsVersion"`
	Platforms   []string                  `json:"platforms,omitempty"`
	SupportedOS []string                  `json:"supportedOS,omitempty"`
	Components  []ComponentRecommendation `json:"components"`
}

// RecommendationResponse is the main API response type
type RecommendationResponse struct {
	Request        RecommendationRequest      `json:"request"`
	MatchedRuleID  string                     `json:"matchedRuleId,omitempty"`
	PayloadVersion string                     `json:"payloadVersion"`
	GeneratedAt    time.Time                  `json:"generatedAt"`
	CNSReleases    []CNSReleaseRecommendation `json:"cnsReleases"`
}

// BulkResolveRequest represents bulk recommendation request
type BulkResolveRequest struct {
	Requests          []RecommendationRequest `json:"requests"`
	PageSize          int                     `json:"pageSize,omitempty"`
	ContinuationToken string                  `json:"continuationToken,omitempty"`
}

// BulkResolveResponse represents bulk recommendation response
type BulkResolveResponse struct {
	Results               []RecommendationResponse `json:"results"`
	NextContinuationToken *string                  `json:"nextContinuationToken"`
	TotalCount            int                      `json:"totalCount,omitempty"`
}

// ErrorResponse represents error responses as per OpenAPI spec
type ErrorResponse struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	RequestID string                 `json:"requestId"`
	Timestamp time.Time              `json:"timestamp"`
	Retryable bool                   `json:"retryable"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason,omitempty"`
}

// Config holds server configuration
type Config struct {
	// Server configuration
	Address string
	Port    int

	// Rate limiting configuration
	RateLimit      rate.Limit // requests per second
	RateLimitBurst int        // burst size

	// Cache configuration
	CacheMaxAge int // seconds

	// Request limits
	MaxBulkRequests int

	// Timeouts
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration

	// Logging
	LogLevel slog.Level
}
