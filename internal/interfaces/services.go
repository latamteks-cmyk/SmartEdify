package interfaces

import (
	"context"

	"github.com/google/uuid"
	"github.com/smartedify/auth-service/internal/models"
)

// UserService interface defines user management operations - aligned with spec requirements
type UserService interface {
	CreateUser(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	GetUserByEmail(ctx context.Context, email, tenantID string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, updates UserUpdates) error
	ValidateCredentials(ctx context.Context, email, password, tenantID string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginInfo LoginInfo) error
	CreatePasswordResetToken(ctx context.Context, email, tenantID string) (string, error)
	ResetPassword(ctx context.Context, token, newPassword string) error
}

// TokenService interface defines token management operations - aligned with spec requirements
type TokenService interface {
	GenerateTokenPair(ctx context.Context, user *models.User) (*models.TokenPair, error)
	ValidateAccessToken(ctx context.Context, token string) (*models.TokenClaims, error)
	ValidateRefreshToken(ctx context.Context, token string) (*models.TokenClaims, error)
	RefreshTokenPair(ctx context.Context, refreshToken string) (*models.TokenPair, error)
	RevokeToken(ctx context.Context, token string) error
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error
}

// SessionService interface defines session management operations - aligned with spec requirements
type SessionService interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	UpdateSession(ctx context.Context, sessionID string, updates SessionUpdates) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
}

// Supporting types for service interfaces
type UserUpdates struct {
	FirstName   *string
	LastName    *string
	Email       *string
	Status      *models.UserStatus
	LastLoginAt *string
	LastLoginIP *string
}

type LoginInfo struct {
	IPAddress string
	UserAgent string
	Timestamp string
}

type SessionUpdates struct {
	LastActivity *string
	IPAddress    *string
	UserAgent    *string
	IsActive     *bool
}