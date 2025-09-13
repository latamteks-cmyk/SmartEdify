package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"

	"github.com/lib/pq"
)

type TokenRepository interface {
	// Refresh token operations
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshToken(ctx context.Context, jti string) (*models.RefreshToken, error)
	UpdateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	RevokeRefreshToken(ctx context.Context, jti, reason string) error
	RevokeAllUserTokens(ctx context.Context, userID, reason string) error
	CleanupExpiredTokens(ctx context.Context) error
	
	// OTP operations
	CreateOTPCode(ctx context.Context, otp *models.OTPCode) error
	GetOTPCode(ctx context.Context, phone, purpose string) (*models.OTPCode, error)
	UseOTPCode(ctx context.Context, id string) error
	IncrementOTPAttempts(ctx context.Context, id string) error
	CleanupExpiredOTPs(ctx context.Context) error
	
	// WebAuthn credentials
	CreateWebAuthnCredential(ctx context.Context, credential *models.WebAuthnCredential) error
	GetWebAuthnCredential(ctx context.Context, credentialID string) (*models.WebAuthnCredential, error)
	GetUserWebAuthnCredentials(ctx context.Context, userID string) ([]*models.WebAuthnCredential, error)
	UpdateWebAuthnCredential(ctx context.Context, credential *models.WebAuthnCredential) error
	DeleteWebAuthnCredential(ctx context.Context, credentialID string) error
	
	// Login attempts tracking
	CreateLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error
	GetRecentLoginAttempts(ctx context.Context, identifier string, since time.Time) ([]*models.LoginAttempt, error)
	CountFailedAttempts(ctx context.Context, identifier string, since time.Time) (int, error)
}

type tokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepository{db: db}
}

// Refresh Token Operations

func (r *tokenRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	deviceInfoJSON, err := json.Marshal(token.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}
	
	query := `
		INSERT INTO refresh_tokens (id, jti, user_id, tenant_id, unit_id, token_hash, expires_at,
		                          device_info, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at`
	
	err = r.db.QueryRowContext(ctx, query,
		token.ID, token.JTI, token.UserID, token.TenantID, token.UnitID,
		token.TokenHash, token.ExpiresAt, deviceInfoJSON, token.IPAddress, token.UserAgent,
	).Scan(&token.CreatedAt, &token.UpdatedAt)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.ErrTokenInvalid.WithDetails("token already exists")
		}
		return fmt.Errorf("failed to create refresh token: %w", err)
	}
	
	return nil
}

func (r *tokenRepository) GetRefreshToken(ctx context.Context, jti string) (*models.RefreshToken, error) {
	token := &models.RefreshToken{}
	var deviceInfoJSON []byte
	
	query := `
		SELECT id, jti, user_id, tenant_id, unit_id, token_hash, expires_at, revoked,
		       revoked_at, revoked_reason, device_info, ip_address, user_agent, created_at, updated_at
		FROM refresh_tokens WHERE jti = $1`
	
	err := r.db.QueryRowContext(ctx, query, jti).Scan(
		&token.ID, &token.JTI, &token.UserID, &token.TenantID, &token.UnitID,
		&token.TokenHash, &token.ExpiresAt, &token.Revoked, &token.RevokedAt,
		&token.RevokedReason, &deviceInfoJSON, &token.IPAddress, &token.UserAgent,
		&token.CreatedAt, &token.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrRefreshTokenInvalid
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	
	if err := json.Unmarshal(deviceInfoJSON, &token.DeviceInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal device info: %w", err)
	}
	
	return token, nil
}

func (r *tokenRepository) UpdateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	deviceInfoJSON, err := json.Marshal(token.DeviceInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal device info: %w", err)
	}
	
	query := `
		UPDATE refresh_tokens 
		SET expires_at = $2, revoked = $3, revoked_at = $4, revoked_reason = $5,
		    device_info = $6, updated_at = NOW()
		WHERE jti = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		token.JTI, token.ExpiresAt, token.Revoked, token.RevokedAt,
		token.RevokedReason, deviceInfoJSON,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update refresh token: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrRefreshTokenInvalid
	}
	
	return nil
}

func (r *tokenRepository) RevokeRefreshToken(ctx context.Context, jti, reason string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked = true, revoked_at = NOW(), revoked_reason = $2, updated_at = NOW()
		WHERE jti = $1`
	
	result, err := r.db.ExecContext(ctx, query, jti, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrRefreshTokenInvalid
	}
	
	return nil
}

func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID, reason string) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked = true, revoked_at = NOW(), revoked_reason = $2, updated_at = NOW()
		WHERE user_id = $1 AND revoked = false`
	
	_, err := r.db.ExecContext(ctx, query, userID, reason)
	if err != nil {
		return fmt.Errorf("failed to revoke all user tokens: %w", err)
	}
	
	return nil
}

func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	
	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	// Log cleanup result (could be 0 if no expired tokens)
	_ = rowsAffected
	
	return nil
}

// OTP Operations

func (r *tokenRepository) CreateOTPCode(ctx context.Context, otp *models.OTPCode) error {
	query := `
		INSERT INTO otp_codes (id, phone, code, code_hash, purpose, expires_at, max_attempts, ip_address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`
	
	err := r.db.QueryRowContext(ctx, query,
		otp.ID, otp.Phone, otp.Code, otp.CodeHash, otp.Purpose,
		otp.ExpiresAt, otp.MaxAttempts, otp.IPAddress,
	).Scan(&otp.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create OTP code: %w", err)
	}
	
	return nil
}

func (r *tokenRepository) GetOTPCode(ctx context.Context, phone, purpose string) (*models.OTPCode, error) {
	otp := &models.OTPCode{}
	
	query := `
		SELECT id, phone, code, code_hash, purpose, expires_at, used, used_at,
		       attempts, max_attempts, ip_address, created_at
		FROM otp_codes 
		WHERE phone = $1 AND purpose = $2 AND used = false AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1`
	
	err := r.db.QueryRowContext(ctx, query, phone, purpose).Scan(
		&otp.ID, &otp.Phone, &otp.Code, &otp.CodeHash, &otp.Purpose,
		&otp.ExpiresAt, &otp.Used, &otp.UsedAt, &otp.Attempts,
		&otp.MaxAttempts, &otp.IPAddress, &otp.CreatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrOTPInvalid
		}
		return nil, fmt.Errorf("failed to get OTP code: %w", err)
	}
	
	return otp, nil
}

func (r *tokenRepository) UseOTPCode(ctx context.Context, id string) error {
	query := `UPDATE otp_codes SET used = true, used_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark OTP as used: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrOTPInvalid
	}
	
	return nil
}

func (r *tokenRepository) IncrementOTPAttempts(ctx context.Context, id string) error {
	query := `UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment OTP attempts: %w", err)
	}
	
	return nil
}

func (r *tokenRepository) CleanupExpiredOTPs(ctx context.Context) error {
	query := `DELETE FROM otp_codes WHERE expires_at < NOW() OR used = true`
	
	_, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired OTPs: %w", err)
	}
	
	return nil
}

// WebAuthn Credentials

func (r *tokenRepository) CreateWebAuthnCredential(ctx context.Context, credential *models.WebAuthnCredential) error {
	flagsJSON, err := json.Marshal(credential.Flags)
	if err != nil {
		return fmt.Errorf("failed to marshal flags: %w", err)
	}
	
	query := `
		INSERT INTO webauthn_credentials (id, user_id, credential_id, public_key, attestation_type,
		                                transport, flags, counter, device_name)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	
	err = r.db.QueryRowContext(ctx, query,
		credential.ID, credential.UserID, credential.CredentialID, credential.PublicKey,
		credential.AttestationType, credential.Transport, flagsJSON, credential.Counter,
		credential.DeviceName,
	).Scan(&credential.CreatedAt, &credential.UpdatedAt)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.ErrCredentialInvalid.WithDetails("credential already exists")
		}
		return fmt.Errorf("failed to create WebAuthn credential: %w", err)
	}
	
	return nil
}

func (r *tokenRepository) GetWebAuthnCredential(ctx context.Context, credentialID string) (*models.WebAuthnCredential, error) {
	credential := &models.WebAuthnCredential{}
	var flagsJSON []byte
	
	query := `
		SELECT id, user_id, credential_id, public_key, attestation_type, transport,
		       flags, counter, device_name, last_used_at, created_at, updated_at
		FROM webauthn_credentials WHERE credential_id = $1`
	
	err := r.db.QueryRowContext(ctx, query, credentialID).Scan(
		&credential.ID, &credential.UserID, &credential.CredentialID, &credential.PublicKey,
		&credential.AttestationType, &credential.Transport, &flagsJSON, &credential.Counter,
		&credential.DeviceName, &credential.LastUsedAt, &credential.CreatedAt, &credential.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrCredentialNotFound
		}
		return nil, fmt.Errorf("failed to get WebAuthn credential: %w", err)
	}
	
	if err := json.Unmarshal(flagsJSON, &credential.Flags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flags: %w", err)
	}
	
	return credential, nil
}

func (r *tokenRepository) GetUserWebAuthnCredentials(ctx context.Context, userID string) ([]*models.WebAuthnCredential, error) {
	query := `
		SELECT id, user_id, credential_id, public_key, attestation_type, transport,
		       flags, counter, device_name, last_used_at, created_at, updated_at
		FROM webauthn_credentials WHERE user_id = $1
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user WebAuthn credentials: %w", err)
	}
	defer rows.Close()
	
	var credentials []*models.WebAuthnCredential
	for rows.Next() {
		credential := &models.WebAuthnCredential{}
		var flagsJSON []byte
		
		err := rows.Scan(
			&credential.ID, &credential.UserID, &credential.CredentialID, &credential.PublicKey,
			&credential.AttestationType, &credential.Transport, &flagsJSON, &credential.Counter,
			&credential.DeviceName, &credential.LastUsedAt, &credential.CreatedAt, &credential.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan WebAuthn credential: %w", err)
		}
		
		if err := json.Unmarshal(flagsJSON, &credential.Flags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal flags: %w", err)
		}
		
		credentials = append(credentials, credential)
	}
	
	return credentials, nil
}

func (r *tokenRepository) UpdateWebAuthnCredential(ctx context.Context, credential *models.WebAuthnCredential) error {
	flagsJSON, err := json.Marshal(credential.Flags)
	if err != nil {
		return fmt.Errorf("failed to marshal flags: %w", err)
	}
	
	query := `
		UPDATE webauthn_credentials 
		SET counter = $2, device_name = $3, last_used_at = $4, flags = $5, updated_at = NOW()
		WHERE credential_id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		credential.CredentialID, credential.Counter, credential.DeviceName,
		credential.LastUsedAt, flagsJSON,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update WebAuthn credential: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrCredentialNotFound
	}
	
	return nil
}

func (r *tokenRepository) DeleteWebAuthnCredential(ctx context.Context, credentialID string) error {
	query := `DELETE FROM webauthn_credentials WHERE credential_id = $1`
	
	result, err := r.db.ExecContext(ctx, query, credentialID)
	if err != nil {
		return fmt.Errorf("failed to delete WebAuthn credential: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrCredentialNotFound
	}
	
	return nil
}

// Login Attempts

func (r *tokenRepository) CreateLoginAttempt(ctx context.Context, attempt *models.LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (id, identifier, identifier_type, attempt_type, success,
		                          ip_address, user_agent, error_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at`
	
	err := r.db.QueryRowContext(ctx, query,
		attempt.ID, attempt.Identifier, attempt.IdentifierType, attempt.AttemptType,
		attempt.Success, attempt.IPAddress, attempt.UserAgent, attempt.ErrorReason,
	).Scan(&attempt.CreatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to create login attempt: %w", err)
	}
	
	return nil
}

func (r *tokenRepository) GetRecentLoginAttempts(ctx context.Context, identifier string, since time.Time) ([]*models.LoginAttempt, error) {
	query := `
		SELECT id, identifier, identifier_type, attempt_type, success, ip_address,
		       user_agent, error_reason, created_at
		FROM login_attempts 
		WHERE identifier = $1 AND created_at >= $2
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, identifier, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent login attempts: %w", err)
	}
	defer rows.Close()
	
	var attempts []*models.LoginAttempt
	for rows.Next() {
		attempt := &models.LoginAttempt{}
		err := rows.Scan(
			&attempt.ID, &attempt.Identifier, &attempt.IdentifierType, &attempt.AttemptType,
			&attempt.Success, &attempt.IPAddress, &attempt.UserAgent, &attempt.ErrorReason,
			&attempt.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan login attempt: %w", err)
		}
		attempts = append(attempts, attempt)
	}
	
	return attempts, nil
}

func (r *tokenRepository) CountFailedAttempts(ctx context.Context, identifier string, since time.Time) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM login_attempts 
		WHERE identifier = $1 AND success = false AND created_at >= $2`
	
	err := r.db.QueryRowContext(ctx, query, identifier, since).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count failed attempts: %w", err)
	}
	
	return count, nil
}