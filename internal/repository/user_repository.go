package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"

	"github.com/lib/pq"
)

type UserRepository interface {
	// User CRUD operations - aligned with spec requirements
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByEmailAndTenant(ctx context.Context, email, tenantID string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateUserStatus(ctx context.Context, userID, status string) error
	UpdateLastLogin(ctx context.Context, userID, ipAddress, userAgent string) error
	IncrementFailedLogins(ctx context.Context, userID string) error
	ResetFailedLogins(ctx context.Context, userID string) error
	LockUser(ctx context.Context, userID string, lockDuration int) error
	
	// Legacy methods for backward compatibility
	GetUserByPhone(ctx context.Context, phone string) (*models.User, error)
	UpdateLastLoginLegacy(ctx context.Context, userID string) error
	
	// User permissions and roles
	GetUserPermissions(ctx context.Context, userID, tenantID, unitID string) (*models.UserPermissions, error)
	GetUserRoles(ctx context.Context, userID string) ([]*models.UserUnitRole, error)
	CreateUserRole(ctx context.Context, role *models.UserUnitRole) error
	UpdateUserRole(ctx context.Context, role *models.UserUnitRole) error
	DeleteUserRole(ctx context.Context, userID, tenantID, unitID string) error
	
	// President management
	GetTenantPresident(ctx context.Context, tenantID string) (*models.TenantPresident, error)
	SetTenantPresident(ctx context.Context, president *models.TenantPresident) error
	GetOwnersByTenant(ctx context.Context, tenantID string) ([]*models.User, error)
	
	// User existence checks
	EmailExists(ctx context.Context, email string) (bool, error)
	PhoneExists(ctx context.Context, phone string) (bool, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, tenant_id, unit_id, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.TenantID, user.UnitID, user.Role, user.Status,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "users_email_key" {
					return errors.ErrEmailAlreadyExists
				}
				if pqErr.Constraint == "users_phone_key" {
					return errors.ErrPhoneAlreadyExists
				}
				return errors.ErrUserAlreadyExists
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}
	
	return nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, first_name, last_name, tenant_id, unit_id, role, status,
		       last_login_at, last_login_ip, failed_logins, locked_until, created_at, updated_at
		FROM users WHERE id = $1`
	
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.TenantID, &user.UnitID, &user.Role, &user.Status,
		&user.LastLoginAt, &user.LastLoginIP, &user.FailedLogins, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	
	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, first_name, last_name, tenant_id, unit_id, role, status,
		       last_login_at, last_login_ip, failed_logins, locked_until, created_at, updated_at
		FROM users WHERE email = $1`
	
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.TenantID, &user.UnitID, &user.Role, &user.Status,
		&user.LastLoginAt, &user.LastLoginIP, &user.FailedLogins, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	
	return user, nil
}

func (r *userRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	// Note: Phone is not in the main users table, this is a legacy method
	// In a real implementation, you might have a separate phone_numbers table
	// For now, return not found as phone is not supported in current schema
	return nil, errors.ErrUserNotFound.WithDetails("phone lookup not supported in current schema")
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, password_hash = $3, first_name = $4, last_name = $5, 
		    tenant_id = $6, unit_id = $7, role = $8, status = $9, updated_at = NOW()
		WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.TenantID, user.UnitID, user.Role, user.Status,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}
	
	return nil
}

func (r *userRepository) UpdateUserStatus(ctx context.Context, userID, status string) error {
	query := `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`
	
	result, err := r.db.ExecContext(ctx, query, userID, status)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrUserNotFound
	}
	
	return nil
}

func (r *userRepository) UpdateLastLoginLegacy(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = NOW(), updated_at = NOW() WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	
	return nil
}

func (r *userRepository) GetUserPermissions(ctx context.Context, userID, tenantID, unitID string) (*models.UserPermissions, error) {
	permissions := &models.UserPermissions{}
	query := `
		SELECT uur.user_id, uur.tenant_id, uur.unit_id, uur.role, uur.is_president, uur.status, u.last_login_at
		FROM user_unit_roles uur
		JOIN users u ON u.id = uur.user_id
		WHERE uur.user_id = $1 AND uur.tenant_id = $2 AND uur.unit_id = $3 AND uur.status = 'active'`
	
	err := r.db.QueryRowContext(ctx, query, userID, tenantID, unitID).Scan(
		&permissions.UserID, &permissions.TenantID, &permissions.UnitID,
		&permissions.Role, &permissions.IsPresident, &permissions.Status,
		&permissions.LastLoginAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrInsufficientScope
		}
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	
	return permissions, nil
}

func (r *userRepository) GetUserRoles(ctx context.Context, userID string) ([]*models.UserUnitRole, error) {
	query := `
		SELECT id, user_id, tenant_id, unit_id, role, is_president, is_primary,
		       percentage_ownership, status, valid_from, valid_until, created_at, updated_at
		FROM user_unit_roles
		WHERE user_id = $1 AND status = 'active'
		ORDER BY created_at DESC`
	
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	defer rows.Close()
	
	var roles []*models.UserUnitRole
	for rows.Next() {
		role := &models.UserUnitRole{}
		err := rows.Scan(
			&role.ID, &role.UserID, &role.TenantID, &role.UnitID,
			&role.Role, &role.IsPresident, &role.IsPrimary,
			&role.PercentageOwnership, &role.Status, &role.ValidFrom,
			&role.ValidUntil, &role.CreatedAt, &role.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user role: %w", err)
		}
		roles = append(roles, role)
	}
	
	return roles, nil
}

func (r *userRepository) CreateUserRole(ctx context.Context, role *models.UserUnitRole) error {
	query := `
		INSERT INTO user_unit_roles (id, user_id, tenant_id, unit_id, role, is_president, is_primary,
		                           percentage_ownership, status, valid_from, valid_until)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`
	
	err := r.db.QueryRowContext(ctx, query,
		role.ID, role.UserID, role.TenantID, role.UnitID, role.Role,
		role.IsPresident, role.IsPrimary, role.PercentageOwnership,
		role.Status, role.ValidFrom, role.ValidUntil,
	).Scan(&role.CreatedAt, &role.UpdatedAt)
	
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return errors.ErrUserAlreadyExists.WithDetails("user already has this role for this unit")
		}
		return fmt.Errorf("failed to create user role: %w", err)
	}
	
	return nil
}

func (r *userRepository) UpdateUserRole(ctx context.Context, role *models.UserUnitRole) error {
	query := `
		UPDATE user_unit_roles 
		SET role = $5, is_president = $6, is_primary = $7, percentage_ownership = $8,
		    status = $9, valid_from = $10, valid_until = $11, updated_at = NOW()
		WHERE user_id = $1 AND tenant_id = $2 AND unit_id = $3 AND role = $4`
	
	result, err := r.db.ExecContext(ctx, query,
		role.UserID, role.TenantID, role.UnitID, role.Role,
		role.Role, role.IsPresident, role.IsPrimary, role.PercentageOwnership,
		role.Status, role.ValidFrom, role.ValidUntil,
	)
	
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrUserNotFound.WithDetails("user role not found")
	}
	
	return nil
}

func (r *userRepository) DeleteUserRole(ctx context.Context, userID, tenantID, unitID string) error {
	query := `DELETE FROM user_unit_roles WHERE user_id = $1 AND tenant_id = $2 AND unit_id = $3`
	
	result, err := r.db.ExecContext(ctx, query, userID, tenantID, unitID)
	if err != nil {
		return fmt.Errorf("failed to delete user role: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return errors.ErrUserNotFound.WithDetails("user role not found")
	}
	
	return nil
}

func (r *userRepository) GetTenantPresident(ctx context.Context, tenantID string) (*models.TenantPresident, error) {
	president := &models.TenantPresident{}
	query := `
		SELECT id, tenant_id, user_id, unit_id, appointed_at, appointed_by,
		       term_start, term_end, status, created_at, updated_at
		FROM tenant_presidents
		WHERE tenant_id = $1 AND status = 'active'`
	
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&president.ID, &president.TenantID, &president.UserID, &president.UnitID,
		&president.AppointedAt, &president.AppointedBy, &president.TermStart,
		&president.TermEnd, &president.Status, &president.CreatedAt, &president.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound.WithDetails("no active president found")
		}
		return nil, fmt.Errorf("failed to get tenant president: %w", err)
	}
	
	return president, nil
}

func (r *userRepository) SetTenantPresident(ctx context.Context, president *models.TenantPresident) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Deactivate current president
	_, err = tx.ExecContext(ctx,
		`UPDATE tenant_presidents SET status = 'inactive', updated_at = NOW() WHERE tenant_id = $1 AND status = 'active'`,
		president.TenantID)
	if err != nil {
		return fmt.Errorf("failed to deactivate current president: %w", err)
	}
	
	// Update user_unit_roles to remove president flag from previous president
	_, err = tx.ExecContext(ctx,
		`UPDATE user_unit_roles SET is_president = false, updated_at = NOW() WHERE tenant_id = $1 AND is_president = true`,
		president.TenantID)
	if err != nil {
		return fmt.Errorf("failed to update previous president role: %w", err)
	}
	
	// Insert new president
	query := `
		INSERT INTO tenant_presidents (id, tenant_id, user_id, unit_id, appointed_at, appointed_by,
		                             term_start, term_end, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at`
	
	err = tx.QueryRowContext(ctx, query,
		president.ID, president.TenantID, president.UserID, president.UnitID,
		president.AppointedAt, president.AppointedBy, president.TermStart,
		president.TermEnd, president.Status,
	).Scan(&president.CreatedAt, &president.UpdatedAt)
	
	if err != nil {
		return fmt.Errorf("failed to insert new president: %w", err)
	}
	
	// Update user_unit_roles to set president flag for new president
	_, err = tx.ExecContext(ctx,
		`UPDATE user_unit_roles SET is_president = true, updated_at = NOW() 
		 WHERE user_id = $1 AND tenant_id = $2 AND unit_id = $3`,
		president.UserID, president.TenantID, president.UnitID)
	if err != nil {
		return fmt.Errorf("failed to update new president role: %w", err)
	}
	
	return tx.Commit()
}

func (r *userRepository) GetOwnersByTenant(ctx context.Context, tenantID string) ([]*models.User, error) {
	query := `
		SELECT DISTINCT u.id, u.email, u.first_name, u.last_name, u.tenant_id, u.unit_id, 
		       u.role, u.status, u.created_at, u.updated_at
		FROM users u
		JOIN user_unit_roles uur ON u.id = uur.user_id
		WHERE uur.tenant_id = $1 AND uur.role = 'owner' AND uur.status = 'active' AND u.status = 'active'
		ORDER BY u.first_name, u.last_name`
	
	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get owners by tenant: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.TenantID, &user.UnitID,
			&user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}
	
	return users, nil
}

// GetUserByEmailAndTenant gets user by email within specific tenant - aligned with spec requirements
func (r *userRepository) GetUserByEmailAndTenant(ctx context.Context, email, tenantID string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, password_hash, first_name, last_name, tenant_id, unit_id, role, status,
		       last_login_at, last_login_ip, failed_logins, locked_until, created_at, updated_at
		FROM users WHERE email = $1 AND tenant_id = $2`
	
	err := r.db.QueryRowContext(ctx, query, email, tenantID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FirstName, &user.LastName,
		&user.TenantID, &user.UnitID, &user.Role, &user.Status,
		&user.LastLoginAt, &user.LastLoginIP, &user.FailedLogins, &user.LockedUntil,
		&user.CreatedAt, &user.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email and tenant: %w", err)
	}
	
	return user, nil
}

// UpdateLastLogin updates user's last login information - aligned with spec requirements
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID, ipAddress, userAgent string) error {
	query := `
		UPDATE users 
		SET last_login_at = NOW(), last_login_ip = $2, updated_at = NOW() 
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	
	return nil
}

// IncrementFailedLogins increments failed login attempts - aligned with spec requirements
func (r *userRepository) IncrementFailedLogins(ctx context.Context, userID string) error {
	query := `
		UPDATE users 
		SET failed_logins = failed_logins + 1, updated_at = NOW() 
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to increment failed logins: %w", err)
	}
	
	return nil
}

// ResetFailedLogins resets failed login attempts - aligned with spec requirements
func (r *userRepository) ResetFailedLogins(ctx context.Context, userID string) error {
	query := `
		UPDATE users 
		SET failed_logins = 0, locked_until = NULL, updated_at = NOW() 
		WHERE id = $1`
	
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to reset failed logins: %w", err)
	}
	
	return nil
}

// LockUser locks user account for specified duration - aligned with spec requirements
func (r *userRepository) LockUser(ctx context.Context, userID string, lockDurationMinutes int) error {
	query := `
		UPDATE users 
		SET status = 'locked', locked_until = NOW() + INTERVAL '%d minutes', updated_at = NOW() 
		WHERE id = $1`
	
	formattedQuery := fmt.Sprintf(query, lockDurationMinutes)
	_, err := r.db.ExecContext(ctx, formattedQuery, userID)
	if err != nil {
		return fmt.Errorf("failed to lock user: %w", err)
	}
	
	return nil
}

func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	
	return exists, nil
}

// EmailExistsInTenant checks if email exists within specific tenant - aligned with spec requirements
func (r *userRepository) EmailExistsInTenant(ctx context.Context, email, tenantID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND tenant_id = $2)`
	
	err := r.db.QueryRowContext(ctx, query, email, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence in tenant: %w", err)
	}
	
	return exists, nil
}

func (r *userRepository) PhoneExists(ctx context.Context, phone string) (bool, error) {
	// Note: Phone is not in the main users table in current schema
	// Always return false for now
	return false, nil
}