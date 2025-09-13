package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system - aligned with spec requirements
type User struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	FirstName    string     `json:"firstName" db:"first_name"`
	LastName     string     `json:"lastName" db:"last_name"`
	TenantID     string     `json:"tenantId" db:"tenant_id"`
	UnitID       string     `json:"unitId" db:"unit_id"`
	Role         string     `json:"role" db:"role"`
	Status       UserStatus `json:"status" db:"status"`
	LastLoginAt  *time.Time `json:"lastLoginAt" db:"last_login_at"`
	LastLoginIP  string     `json:"lastLoginIp" db:"last_login_ip"`
	FailedLogins int        `json:"-" db:"failed_logins"`
	LockedUntil  *time.Time `json:"-" db:"locked_until"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time  `json:"updatedAt" db:"updated_at"`
	
	// Legacy fields for backward compatibility (not in main users table)
	Phone         string     `json:"phone,omitempty"`
	Name          string     `json:"name,omitempty"`
	EmailVerified bool       `json:"emailVerified,omitempty"`
	PhoneVerified bool       `json:"phoneVerified,omitempty"`
	MFAEnabled    bool       `json:"mfaEnabled,omitempty"`
	MFASecret     string     `json:"-"`
}

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusLocked    UserStatus = "locked"
)

// UserUnitRole represents the relationship between a user and a unit with a specific role
type UserUnitRole struct {
	ID                   string     `json:"id" db:"id"`
	UserID               string     `json:"user_id" db:"user_id"`
	TenantID             string     `json:"tenant_id" db:"tenant_id"`
	UnitID               string     `json:"unit_id" db:"unit_id"`
	Role                 string     `json:"role" db:"role"`
	IsPresident          bool       `json:"is_president" db:"is_president"`
	IsPrimary            bool       `json:"is_primary" db:"is_primary"`
	PercentageOwnership  float64    `json:"percentage_ownership" db:"percentage_ownership"`
	Status               string     `json:"status" db:"status"`
	ValidFrom            time.Time  `json:"valid_from" db:"valid_from"`
	ValidUntil           *time.Time `json:"valid_until" db:"valid_until"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// Tenant represents a condominium or property management entity
type Tenant struct {
	ID        string                 `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	LegalName *string                `json:"legal_name" db:"legal_name"`
	TaxID     *string                `json:"tax_id" db:"tax_id"`
	Address   *string                `json:"address" db:"address"`
	City      *string                `json:"city" db:"city"`
	State     *string                `json:"state" db:"state"`
	Country   string                 `json:"country" db:"country"`
	Phone     *string                `json:"phone" db:"phone"`
	Email     *string                `json:"email" db:"email"`
	Status    string                 `json:"status" db:"status"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// Unit represents a property unit within a tenant
type Unit struct {
	ID           string                 `json:"id" db:"id"`
	TenantID     string                 `json:"tenant_id" db:"tenant_id"`
	UnitNumber   string                 `json:"unit_number" db:"unit_number"`
	UnitType     string                 `json:"unit_type" db:"unit_type"`
	FloorNumber  *int                   `json:"floor_number" db:"floor_number"`
	Building     *string                `json:"building" db:"building"`
	AreaSqm      *float64               `json:"area_sqm" db:"area_sqm"`
	Bedrooms     *int                   `json:"bedrooms" db:"bedrooms"`
	Bathrooms    *int                   `json:"bathrooms" db:"bathrooms"`
	Status       string                 `json:"status" db:"status"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at" db:"updated_at"`
}

// TenantPresident represents the current president of a tenant
type TenantPresident struct {
	ID          string     `json:"id" db:"id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	UserID      string     `json:"user_id" db:"user_id"`
	UnitID      string     `json:"unit_id" db:"unit_id"`
	AppointedAt time.Time  `json:"appointed_at" db:"appointed_at"`
	AppointedBy *string    `json:"appointed_by" db:"appointed_by"`
	TermStart   time.Time  `json:"term_start" db:"term_start"`
	TermEnd     *time.Time `json:"term_end" db:"term_end"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Session represents a user session - aligned with spec requirements
type Session struct {
	ID           string    `json:"id" redis:"id"`
	UserID       uuid.UUID `json:"userId" redis:"user_id"`
	TenantID     string    `json:"tenantId" redis:"tenant_id"`
	AccessToken  string    `json:"-" redis:"access_token"`
	RefreshToken string    `json:"-" redis:"refresh_token"`
	ExpiresAt    time.Time `json:"expiresAt" redis:"expires_at"`
	CreatedAt    time.Time `json:"createdAt" redis:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" redis:"updated_at"`
	IPAddress    string    `json:"ipAddress" redis:"ip_address"`
	UserAgent    string    `json:"userAgent" redis:"user_agent"`
	IsActive     bool      `json:"isActive" redis:"is_active"`
}

// TokenPair represents JWT token pair - aligned with spec requirements
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int    `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

// TokenClaims represents JWT token claims - aligned with spec requirements
type TokenClaims struct {
	UserID      uuid.UUID `json:"sub"`
	Email       string    `json:"email"`
	TenantID    string    `json:"tenant_id"`
	UnitID      string    `json:"unit_id"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	SessionID   string    `json:"session_id"`
	// Standard JWT claims will be embedded
	Issuer    string `json:"iss"`
	Audience  string `json:"aud"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	ID            string                 `json:"id" db:"id"`
	JTI           string                 `json:"jti" db:"jti"`
	UserID        string                 `json:"user_id" db:"user_id"`
	TenantID      *string                `json:"tenant_id" db:"tenant_id"`
	UnitID        *string                `json:"unit_id" db:"unit_id"`
	TokenHash     string                 `json:"-" db:"token_hash"`
	ExpiresAt     time.Time              `json:"expires_at" db:"expires_at"`
	Revoked       bool                   `json:"revoked" db:"revoked"`
	RevokedAt     *time.Time             `json:"revoked_at" db:"revoked_at"`
	RevokedReason *string                `json:"revoked_reason" db:"revoked_reason"`
	DeviceInfo    map[string]interface{} `json:"device_info" db:"device_info"`
	IPAddress     *string                `json:"ip_address" db:"ip_address"`
	UserAgent     *string                `json:"user_agent" db:"user_agent"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// WebAuthnCredential represents a WebAuthn credential
type WebAuthnCredential struct {
	ID              string                 `json:"id" db:"id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	CredentialID    string                 `json:"credential_id" db:"credential_id"`
	PublicKey       string                 `json:"public_key" db:"public_key"`
	AttestationType *string                `json:"attestation_type" db:"attestation_type"`
	Transport       *string                `json:"transport" db:"transport"`
	Flags           map[string]interface{} `json:"flags" db:"flags"`
	Counter         int64                  `json:"counter" db:"counter"`
	DeviceName      *string                `json:"device_name" db:"device_name"`
	LastUsedAt      *time.Time             `json:"last_used_at" db:"last_used_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// LoginAttempt represents a login attempt for security tracking
type LoginAttempt struct {
	ID             string     `json:"id" db:"id"`
	Identifier     string     `json:"identifier" db:"identifier"`
	IdentifierType string     `json:"identifier_type" db:"identifier_type"`
	AttemptType    string     `json:"attempt_type" db:"attempt_type"`
	Success        bool       `json:"success" db:"success"`
	IPAddress      *string    `json:"ip_address" db:"ip_address"`
	UserAgent      *string    `json:"user_agent" db:"user_agent"`
	ErrorReason    *string    `json:"error_reason" db:"error_reason"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// OTPCode represents an OTP code for verification
type OTPCode struct {
	ID          string    `json:"id" db:"id"`
	Phone       string    `json:"phone" db:"phone"`
	Code        string    `json:"-" db:"code"`
	CodeHash    string    `json:"-" db:"code_hash"`
	Purpose     string    `json:"purpose" db:"purpose"`
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	Used        bool      `json:"used" db:"used"`
	UsedAt      *time.Time `json:"used_at" db:"used_at"`
	Attempts    int       `json:"attempts" db:"attempts"`
	MaxAttempts int       `json:"max_attempts" db:"max_attempts"`
	IPAddress   *string   `json:"ip_address" db:"ip_address"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID            string                 `json:"id" db:"id"`
	EventType     string                 `json:"event_type" db:"event_type"`
	EventCategory string                 `json:"event_category" db:"event_category"`
	UserID        *string                `json:"user_id" db:"user_id"`
	TenantID      *string                `json:"tenant_id" db:"tenant_id"`
	UnitID        *string                `json:"unit_id" db:"unit_id"`
	ResourceType  *string                `json:"resource_type" db:"resource_type"`
	ResourceID    *string                `json:"resource_id" db:"resource_id"`
	Action        string                 `json:"action" db:"action"`
	Details       map[string]interface{} `json:"details" db:"details"`
	IPAddress     *string                `json:"ip_address" db:"ip_address"`
	UserAgent     *string                `json:"user_agent" db:"user_agent"`
	SessionID     *string                `json:"session_id" db:"session_id"`
	CorrelationID *string                `json:"correlation_id" db:"correlation_id"`
	HashChain     *string                `json:"hash_chain" db:"hash_chain"`
	PreviousHash  *string                `json:"previous_hash" db:"previous_hash"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// UserPermissions represents a user's permissions for a specific context
type UserPermissions struct {
	UserID      string    `json:"user_id"`
	TenantID    string    `json:"tenant_id"`
	UnitID      string    `json:"unit_id"`
	Role        string    `json:"role"`
	IsPresident bool      `json:"is_president"`
	Status      string    `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at"`
}

// RegisterRequest represents a user registration request - aligned with spec requirements
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required,min=2,max=50"`
	LastName  string `json:"lastName" validate:"required,min=2,max=50"`
	TenantID  string `json:"tenantId" validate:"required"`
	UnitID    string `json:"unitId" validate:"required"`
}

// LoginRequest represents a login request - aligned with spec requirements
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenantId" validate:"required"`
	UnitID   string `json:"unitId" validate:"required"`
}

// AuthResponse represents authentication response - aligned with spec requirements
type AuthResponse struct {
	User          *User      `json:"user"`
	TokenPair     *TokenPair `json:"tokens"`
	SessionID     string     `json:"sessionId"`
	CorrelationID string     `json:"correlationId"`
}

// CreateUserRequest represents a request to create a new user
type CreateUserRequest struct {
	Email         string `json:"email" validate:"required,email"`
	Phone         string `json:"phone" validate:"required"`
	Name          string `json:"name" validate:"required"`
	Password      string `json:"password,omitempty" validate:"omitempty,min=12"`
	TenantID      string `json:"tenant_id" validate:"required,uuid"`
	TermsAccepted bool   `json:"terms_accepted" validate:"required"`
}

// WhatsAppLoginRequest represents a WhatsApp login request
type WhatsAppLoginRequest struct {
	Phone string `json:"phone" validate:"required"`
	OTP   string `json:"otp" validate:"required"`
}

// LoginResponse represents a successful login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	UserID       string `json:"user_id"`
	TenantID     string `json:"tenant_id,omitempty"`
	UnitID       string `json:"unit_id,omitempty"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TransferPresidentRequest represents a request to transfer presidency
type TransferPresidentRequest struct {
	ToUserID string `json:"to_user_id" validate:"required,uuid"`
}

// ARCORequest represents an ARCO rights request
type ARCORequest struct {
	RequestType string                 `json:"request_type" validate:"required,oneof=access rectify delete"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Reason      string                 `json:"reason,omitempty"`
}