package server

// Error codes as constants
const (
	ErrCodeInvalidParameter   = "INVALID_PARAMETER"
	ErrCodeInvalidJSON        = "INVALID_JSON"
	ErrCodeMethodNotAllowed   = "METHOD_NOT_ALLOWED"
	ErrCodeNoMatchingRule     = "NO_MATCHING_RULE"
	ErrCodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrCodeInternalError      = "INTERNAL_ERROR"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)
