package server

import (
	"fmt"
	"regexp"
)

// Validation patterns and constants

var (
	// Compiled regex patterns from OpenAPI spec
	osVersionPattern      = regexp.MustCompile(`^(ALL|\d+\.\d+)$`)
	kernelPattern         = regexp.MustCompile(`^(ALL|\d+\.\d+(\.\d+)?)$`)
	kubernetesPattern     = regexp.MustCompile(`^(ALL|1\.\d+)$`)
	payloadVersionPattern = regexp.MustCompile(`^\d{4}\.\d{1,2}\.\d+$`)
)

// Valid enum values from OpenAPI spec
var (
	validOSFamilies   = []string{"Ubuntu", "RHEL", "ALL"}
	validEnvironments = []string{"GKE", "EKS", "OKE", "ALL"}
	validGPUs         = []string{"H100", "GB200", "A100", "L40", "ALL"}
	validIntents      = []string{"training", "inference", "ALL"}
)

// Validator handles request validation
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRecommendationRequest validates a recommendation request according to OpenAPI spec
func (v *Validator) ValidateRecommendationRequest(req *RecommendationRequest) error {
	// Validate osFamily enum
	if !isValidEnum(req.OSFamily, validOSFamilies) {
		return fmt.Errorf("invalid osFamily: must be one of %v", validOSFamilies)
	}

	// Validate osVersion pattern
	if !osVersionPattern.MatchString(req.OSVersion) {
		return fmt.Errorf("invalid osVersion format: must match pattern (ALL|\\d+\\.\\d+)")
	}

	// Validate kernel pattern
	if !kernelPattern.MatchString(req.Kernel) {
		return fmt.Errorf("invalid kernel format: must match pattern (ALL|\\d+\\.\\d+(\\.\\d+)?)")
	}

	// Validate environment enum
	if !isValidEnum(req.Environment, validEnvironments) {
		return fmt.Errorf("invalid environment: must be one of %v", validEnvironments)
	}

	// Validate kubernetes pattern
	if !kubernetesPattern.MatchString(req.Kubernetes) {
		return fmt.Errorf("invalid kubernetes format: must match pattern (ALL|1\\.\\d+)")
	}

	// Validate gpu enum
	if !isValidEnum(req.GPU, validGPUs) {
		return fmt.Errorf("invalid gpu: must be one of %v", validGPUs)
	}

	// Validate intent enum
	if !isValidEnum(req.Intent, validIntents) {
		return fmt.Errorf("invalid intent: must be one of %v", validIntents)
	}

	// Validate payloadVersion if provided
	if req.PayloadVersionRequested != nil && *req.PayloadVersionRequested != "" {
		if !payloadVersionPattern.MatchString(*req.PayloadVersionRequested) {
			return fmt.Errorf("invalid payloadVersion format: must match pattern \\d{4}\\.\\d{1,2}\\.\\d+")
		}
	}

	return nil
}

// isValidEnum checks if a value exists in a list of valid values
func isValidEnum(value string, validValues []string) bool {
	for _, v := range validValues {
		if value == v {
			return true
		}
	}
	return false
}
