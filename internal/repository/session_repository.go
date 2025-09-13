package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"
)

// SessionRepository interface defines session management operations - aligned with spec requirements
type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.Session) error
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	UpdateSession(ctx context.Context, sessionID string, updates SessionUpdates) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredSessions(ctx context.Context) error
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error)
}

type SessionUpdates struct {
	LastActivity *time.Time
	IPAddress    *string
	UserAgent    *string
	IsActive     *bool
}

type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, tenant_id, access_token_hash, refresh_token_hash, 
		                     expires_at, ip_address, user_agent, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		session.ID, session.UserID, session.TenantID, session.AccessToken,
		session.RefreshToken, session.ExpiresAt, session.IPAddress,
		session.UserAgent, session.IsActive,
	).Scan(&session.CreatedAt, &session.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	
	return nil
}

func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session := &models.Session{}
	query := `
		SELECT id, user_id, tenant_id, access_token_hash, refresh_token_hash,
		       expires_at, created_at, updated_at, ip_address, user_agent, is_active
		FROM sessions WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.AccessToken,
		&session.RefreshToken, &session.ExpiresAt, &session.CreatedAt,
		&session.UpdatedAt, &session.IPAddress, &session.UserAgent, &session.IsActive,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return session, nil
}

func (r *sessionRepository) UpdateSession(ctx context.Context, sessionID string, updates SessionUpdates) error {
	query := `
		UPDATE sessions 
		SET updated_at = NOW()`
	
	args := []interface{}{sessionID}
	argIndex := 2
	
	if updates.LastActivity != nil {
		query += fmt.Sprintf(", updated_at = $%d", argIndex)
		args = append(args, *updates.LastActivity)
		argIndex++
	}
	
	if updates.IPAddress != nil {
		query += fmt.Sprintf(", ip_address = $%d", argIndex)
		args = append(args, *updates.IPAddress)
		argIndex++
	}
	
	if updates.UserAgent != nil {
		query += fmt.Sprintf(", user_agent = $%d", argIndex)
		args = append(args, *updates.UserAgent)
		argIndex++
	}
	
	if updates.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argIndex)
		args = append(args, *updates.IsActive)
		argIndex++
	}
	
	query += " WHERE id = $1"
	
	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrSessionNotFound
	}
	
	return nil
}

func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrSessionNotFound
	}
	
	return nil
}

func (r *sessionRepository) DeleteAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}
	
	return nil
}

func (r *sessionRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`
	
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	// Log the number of cleaned up sessions (could be 0)
	_ = rowsAffected
	
	return nil
}

func (r *sessionRepository) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*models.Session, error) {
	query := `
		SELECT id, user_id, tenant_id, access_token_hash, refresh_token_hash,
		       expires_at, created_at, updated_at, ip_address, user_agent, is_active
		FROM sessions 
		WHERE user_id = $1 AND is_active = true AND expires_at > NOW()
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}
	defer rows.Close()
	
	var sessions []*models.Session
	for rows.Next() {
		session := &models.Session{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.TenantID, &session.AccessToken,
			&session.RefreshToken, &session.ExpiresAt, &session.CreatedAt,
			&session.UpdatedAt, &session.IPAddress, &session.UserAgent, &session.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}
	
	return sessions, nil
}