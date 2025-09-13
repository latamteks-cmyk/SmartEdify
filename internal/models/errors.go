package models

import "time"

// ErrorResponse represents standardized error response format - aligned with spec requirements
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code          string            `json:"code"`
	Message       string            `json:"message"`
	Details       map[string]string `json:"details,omitempty"`
	CorrelationID string            `json:"correlationId"`
	Timestamp     time.Time         `json:"timestamp"`
}

// Error codes mapping to HTTP status codes - aligned with spec requirements
var ErrorCodeMapping = map[string]int{
	"EMAIL_ALREADY_EXISTS":     409,
	"INVALID_EMAIL_FORMAT":     400,
	"WEAK_PASSWORD":           400,
	"MISSING_REQUIRED_FIELDS": 400,
	"INVALID_CREDENTIALS":     401,
	"TOKEN_EXPIRED":           401,
	"TOKEN_INVALID":           401,
	"ACCOUNT_SUSPENDED":       403,
	"INVALID_TENANT_ACCESS":   403,
	"EMAIL_NOT_VERIFIED":      403,
	"RATE_LIMIT_EXCEEDED":     429,
	"INTERNAL_SERVER_ERROR":   500,
}