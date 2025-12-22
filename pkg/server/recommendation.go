package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Handler implementations

// handleGetRecommendations handles GET /v1/recommendations
func (s *Server) handleGetRecommendations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, r, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed,
			"Method not allowed", false, nil)
		return
	}

	// Parse and validate query parameters
	req := s.parseRecommendationRequest(r)
	if err := s.validator.ValidateRecommendationRequest(req); err != nil {
		s.writeError(w, r, http.StatusBadRequest, ErrCodeInvalidParameter,
			err.Error(), false, map[string]interface{}{
				"request": req,
			})
		return
	}

	// Generate recommendation (stub implementation)
	resp := s.generateRecommendation(req)

	// Set cache headers
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", s.config.CacheMaxAge))

	s.writeJSON(w, http.StatusOK, resp)
}

// handleBulkResolve handles POST /v1/recommendations/resolve
func (s *Server) handleBulkResolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, r, http.StatusMethodNotAllowed, ErrCodeMethodNotAllowed,
			"Method not allowed", false, nil)
		return
	}

	var bulkReq BulkResolveRequest
	if err := json.NewDecoder(r.Body).Decode(&bulkReq); err != nil {
		s.writeError(w, r, http.StatusBadRequest, ErrCodeInvalidJSON,
			"Invalid JSON payload", false, map[string]interface{}{
				"error": err.Error(),
			})
		return
	}

	// Validate bulk request
	if len(bulkReq.Requests) == 0 {
		s.writeError(w, r, http.StatusBadRequest, ErrCodeInvalidParameter,
			"requests array cannot be empty", false, nil)
		return
	}

	if len(bulkReq.Requests) > s.config.MaxBulkRequests {
		s.writeError(w, r, http.StatusBadRequest, ErrCodeInvalidParameter,
			fmt.Sprintf("requests array exceeds maximum of %d", s.config.MaxBulkRequests),
			false, map[string]interface{}{
				"maxAllowed": s.config.MaxBulkRequests,
				"provided":   len(bulkReq.Requests),
			})
		return
	}

	// Validate each request
	for i, req := range bulkReq.Requests {
		if err := s.validator.ValidateRecommendationRequest(&req); err != nil {
			s.writeError(w, r, http.StatusBadRequest, ErrCodeInvalidParameter,
				fmt.Sprintf("invalid request at index %d: %s", i, err.Error()),
				false, map[string]interface{}{
					"index":   i,
					"request": req,
				})
			return
		}
	}

	// Process bulk request
	results := make([]RecommendationResponse, len(bulkReq.Requests))
	for i, req := range bulkReq.Requests {
		results[i] = s.generateRecommendation(&req)
	}

	resp := BulkResolveResponse{
		Results:    results,
		TotalCount: len(results),
	}

	s.writeJSON(w, http.StatusOK, resp)
}

// Helper methods

// parseRecommendationRequest extracts query parameters into request struct
func (s *Server) parseRecommendationRequest(r *http.Request) *RecommendationRequest {
	q := r.URL.Query()

	req := &RecommendationRequest{
		OSFamily:    getQueryParamOrDefault(q, "osFamily"),
		OSVersion:   getQueryParamOrDefault(q, "osVersion"),
		Kernel:      getQueryParamOrDefault(q, "kernel"),
		Environment: getQueryParamOrDefault(q, "environment"),
		Kubernetes:  getQueryParamOrDefault(q, "kubernetes"),
		GPU:         getQueryParamOrDefault(q, "gpu"),
		Intent:      getQueryParamOrDefault(q, "intent"),
	}

	if pv := q.Get("payloadVersion"); pv != "" {
		req.PayloadVersionRequested = &pv
	}

	return req
}

// generateRecommendation creates a recommendation response (stub implementation)
func (s *Server) generateRecommendation(req *RecommendationRequest) RecommendationResponse {
	// This is a stub implementation. In production, this would:
	// 1. Query a database or rule engine
	// 2. Match rules based on request parameters
	// 3. Return appropriate CNS release recommendations

	version := "2025.12.0"
	components := []ComponentRecommendation{
		{Name: "containerd", Version: stringPtr("2.1.3")},
		{Name: "nvidia-container-toolkit", Version: stringPtr("1.17.8")},
		{Name: "kubernetes", Version: stringPtr("1.33.2")},
		{Name: "nvidia-gpu-operator", Version: stringPtr("25.3.4")},
		{Name: "nvidia-data-center-driver", Version: stringPtr("580.82.07")},
	}

	return RecommendationResponse{
		Request:        *req,
		MatchedRuleID:  "default-rule",
		PayloadVersion: version,
		GeneratedAt:    time.Now().UTC(),
		CNSReleases: []CNSReleaseRecommendation{
			{
				CNSVersion:  "16.0",
				Platforms:   []string{"NVIDIA Certified Server (x86 & arm64)", "DGX Server"},
				SupportedOS: []string{"Ubuntu 24.04 LTS"},
				Components:  components,
			},
		},
	}
}
