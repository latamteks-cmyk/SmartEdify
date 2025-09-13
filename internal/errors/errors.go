package errors

import (
	"fmt"
	"net/http"
)

// Custom error types
var (
	// Authentication errors
	ErrUserNotFound        = NewAPIError("USER_NOT_FOUND", "User not found", http.StatusNotFound)
	ErrInvalidCredentials  = NewAPIError("INVALID_CREDENTIALS", "Invalid credentials", http.StatusUnauthorized)
	ErrTokenExpired        = NewAPIError("TOKEN_EXPIRED", "Token has expired", http.StatusUnauthorized)
	ErrTokenInvalid        = NewAPIError("TOKEN_INVALID", "Invalid token", http.StatusUnauthorized)
	ErrRefreshTokenInvalid = NewAPIError("REFRESH_TOKEN_INVALID", "Invalid refresh token", http.StatusUnauthorized)
	ErrTokenRevoked        = NewAPIError("TOKEN_REVOKED", "Token has been revoked", http.StatusUnauthorized)

	// Authorization errors
	ErrInsufficientScope   = NewAPIError("INSUFFICIENT_SCOPE", "Insufficient permissions", http.StatusForbidden)
	ErrUnauthorizedAccess  = NewAPIError("UNAUTHORIZED_ACCESS", "Unauthorized access", http.StatusForbidden)
	ErrTenantAccessDenied  = NewAPIError("TENANT_ACCESS_DENIED", "Access denied for this tenant", http.StatusForbidden)
	ErrUnitAccessDenied    = NewAPIError("UNIT_ACCESS_DENIED", "Access denied for this unit", http.StatusForbidden)

	// Validation errors
	ErrValidationFailed    = NewAPIError("VALIDATION_FAILED", "Validation failed", http.StatusBadRequest)
	ErrInvalidInput        = NewAPIError("INVALID_INPUT", "Invalid input provided", http.StatusBadRequest)
	ErrMissingRequired     = NewAPIError("MISSING_REQUIRED", "Required field is missing", http.StatusBadRequest)
	ErrInvalidFormat       = NewAPIError("INVALID_FORMAT", "Invalid format", http.StatusBadRequest)
	ErrWeakPassword       = NewAPIError("WEAK_PASSWORD", "La contraseña no cumple los requisitos de seguridad", http.StatusBadRequest)

	// Rate limiting errors
	ErrRateLimitExceeded   = NewAPIError("RATE_LIMIT_EXCEEDED", "Rate limit exceeded", http.StatusTooManyRequests)
	ErrTooManyAttempts     = NewAPIError("TOO_MANY_ATTEMPTS", "Too many attempts", http.StatusTooManyRequests)
	ErrAccountLocked       = NewAPIError("ACCOUNT_LOCKED", "Account temporarily locked", http.StatusTooManyRequests)

	// Security errors
	ErrInvalidDPoP         = NewAPIError("INVALID_DPOP", "Invalid DPoP proof", http.StatusBadRequest)
	ErrDPoPReplay          = NewAPIError("DPOP_REPLAY", "DPoP proof replay detected", http.StatusBadRequest)
	ErrInvalidSignature    = NewAPIError("INVALID_SIGNATURE", "Invalid signature", http.StatusBadRequest)
	ErrEncryptionFailed    = NewAPIError("ENCRYPTION_FAILED", "Encryption failed", http.StatusInternalServerError)
	ErrDecryptionFailed    = NewAPIError("DECRYPTION_FAILED", "Decryption failed", http.StatusInternalServerError)

	// OTP errors
	ErrOTPExpired          = NewAPIError("OTP_EXPIRED", "OTP has expired", http.StatusBadRequest)
	ErrOTPInvalid          = NewAPIError("OTP_INVALID", "Invalid OTP", http.StatusBadRequest)
	ErrOTPAlreadyUsed      = NewAPIError("OTP_ALREADY_USED", "OTP has already been used", http.StatusBadRequest)
	ErrOTPMaxAttempts      = NewAPIError("OTP_MAX_ATTEMPTS", "Maximum OTP attempts exceeded", http.StatusTooManyRequests)

	// WebAuthn errors
	ErrWebAuthnFailed      = NewAPIError("WEBAUTHN_FAILED", "WebAuthn authentication failed", http.StatusBadRequest)
	ErrCredentialNotFound  = NewAPIError("CREDENTIAL_NOT_FOUND", "Credential not found", http.StatusNotFound)
	ErrCredentialInvalid   = NewAPIError("CREDENTIAL_INVALID", "Invalid credential", http.StatusBadRequest)

	// Business logic errors
	ErrUserAlreadyExists   = NewAPIError("USER_ALREADY_EXISTS", "User already exists", http.StatusConflict)
	ErrEmailAlreadyExists  = NewAPIError("EMAIL_ALREADY_EXISTS", "Email already exists", http.StatusConflict)
	ErrPhoneAlreadyExists  = NewAPIError("PHONE_ALREADY_EXISTS", "Phone already exists", http.StatusConflict)
	ErrUserNotActive       = NewAPIError("USER_NOT_ACTIVE", "User account is not active", http.StatusForbidden)
	ErrUserSuspended       = NewAPIError("USER_SUSPENDED", "User account is suspended", http.StatusForbidden)

	// President/Role errors
	ErrNotOwner            = NewAPIError("NOT_OWNER", "Only property owners can be president", http.StatusForbidden)
	ErrAlreadyPresident    = NewAPIError("ALREADY_PRESIDENT", "User is already president", http.StatusConflict)
	ErrPresidentRequired   = NewAPIError("PRESIDENT_REQUIRED", "President role required for this action", http.StatusForbidden)

	// ARCO errors
	ErrARCORequestFailed   = NewAPIError("ARCO_REQUEST_FAILED", "ARCO request failed", http.StatusBadRequest)
	ErrARCONotAuthorized   = NewAPIError("ARCO_NOT_AUTHORIZED", "Not authorized for ARCO request", http.StatusForbidden)
	ErrARCOMFARequired     = NewAPIError("ARCO_MFA_REQUIRED", "MFA required for ARCO request", http.StatusUnauthorized)

	// External service errors
	ErrWhatsAppFailed      = NewAPIError("WHATSAPP_FAILED", "WhatsApp service failed", http.StatusServiceUnavailable)
	ErrHSMFailed           = NewAPIError("HSM_FAILED", "HSM service failed", http.StatusServiceUnavailable)
	ErrComplianceFailed    = NewAPIError("COMPLIANCE_FAILED", "Compliance service failed", http.StatusServiceUnavailable)
	ErrIPFSFailed          = NewAPIError("IPFS_FAILED", "IPFS service failed", http.StatusServiceUnavailable)

	// Session errors
	ErrSessionNotFound     = NewAPIError("SESSION_NOT_FOUND", "Session not found", http.StatusNotFound)
	ErrSessionExpired      = NewAPIError("SESSION_EXPIRED", "Session has expired", http.StatusUnauthorized)
	ErrSessionInvalid      = NewAPIError("SESSION_INVALID", "Invalid session", http.StatusUnauthorized)

	// Cache errors
	ErrCacheMiss           = NewAPIError("CACHE_MISS", "Cache miss", http.StatusNotFound)
	ErrCacheWriteFailed    = NewAPIError("CACHE_WRITE_FAILED", "Cache write failed", http.StatusInternalServerError)

	// Database errors
	ErrDatabaseConnection  = NewAPIError("DATABASE_CONNECTION", "Database connection failed", http.StatusServiceUnavailable)
	ErrDatabaseQuery       = NewAPIError("DATABASE_QUERY", "Database query failed", http.StatusInternalServerError)
	ErrDatabaseTransaction = NewAPIError("DATABASE_TRANSACTION", "Database transaction failed", http.StatusInternalServerError)

	// Generic errors
	ErrInternalServer      = NewAPIError("INTERNAL_SERVER_ERROR", "Internal server error", http.StatusInternalServerError)
	ErrServiceUnavailable  = NewAPIError("SERVICE_UNAVAILABLE", "Service unavailable", http.StatusServiceUnavailable)
	ErrTimeout             = NewAPIError("TIMEOUT", "Request timeout", http.StatusRequestTimeout)
)

// APIError represents a structured API error
type APIError struct {
	Code    string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Status  int         `json:"-"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *APIError) WithDetails(details interface{}) *APIError {
	return &APIError{
		Code:    e.Code,
		Message: e.Message,
		Details: details,
		Status:  e.Status,
	}
}

func (e *APIError) WithMessage(message string) *APIError {
	return &APIError{
		Code:    e.Code,
		Message: message,
		Details: e.Details,
		Status:  e.Status,
	}
}

// NewAPIError creates a new API error
func NewAPIError(code, message string, status int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// ValidationError represents field validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	return fmt.Sprintf("validation failed: %s", ve[0].Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// NewValidationErrors creates validation errors from a map
func NewValidationErrors(errors map[string]string) ValidationErrors {
	var validationErrors ValidationErrors
	for field, message := range errors {
		validationErrors = append(validationErrors, ValidationError{
			Field:   field,
			Message: message,
		})
	}
	return validationErrors
}

// IsAPIError checks if an error is an APIError
func IsAPIError(err error) (*APIError, bool) {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr, true
	}
	return nil, false
}

// WrapError wraps a generic error as an internal server error
func WrapError(err error) *APIError {
	if apiErr, ok := IsAPIError(err); ok {
		return apiErr
	}
	return ErrInternalServer.WithDetails(err.Error())
}