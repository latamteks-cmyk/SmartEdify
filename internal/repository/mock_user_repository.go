package repository

import (
	"context"
	"time"

	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"
)

// Obtener usuario por email y tenant
func (r *MockUserRepository) GetUserByEmailAndTenant(ctx context.Context, email, tenantID string) (*models.User, error) {
	key := email + ":" + tenantID
	userID, exists := r.DB.EmailTenantIndex[key]
	if !exists {
		return nil, errors.ErrUserNotFound
	}
	user, ok := r.DB.Users[userID]
	if !ok {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

// Mock database for testing
// Exported so tests in other packages can use
// (no testing imports here)
type MockDB struct {
	Users            map[string]*models.User
	EmailTenantIndex map[string]string // email+tenant -> userID
}

func NewMockDB() *MockDB {
	return &MockDB{
		Users:            make(map[string]*models.User),
		EmailTenantIndex: make(map[string]string),
	}
}

// Mock repository for testing
// Exported so tests in other packages can use
// Implements the same interface as real repository
// (no testing imports here)
type MockUserRepository struct {
	DB *MockDB
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		DB: NewMockDB(),
	}
}

func (r *MockUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	emailTenantKey := user.Email + ":" + user.TenantID
	if _, exists := r.DB.EmailTenantIndex[emailTenantKey]; exists {
		return errors.ErrEmailAlreadyExists
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	userID := user.ID.String()
	r.DB.Users[userID] = user
	r.DB.EmailTenantIndex[emailTenantKey] = userID
	return nil
}

func (r *MockUserRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, exists := r.DB.Users[id]
	if !exists {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (r *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	for _, user := range r.DB.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.ErrUserNotFound
}

// Actualiza la información de último login del usuario
func (r *MockUserRepository) UpdateLastLogin(ctx context.Context, userID, ipAddress, userAgent string) error {
	user, exists := r.DB.Users[userID]
	if !exists {
		return errors.ErrUserNotFound
	}
	now := time.Now()
	user.LastLoginAt = &now
	user.LastLoginIP = ipAddress
	user.UpdatedAt = now
	return nil
}

// Incrementa el contador de intentos fallidos de login
func (r *MockUserRepository) IncrementFailedLogins(ctx context.Context, userID string) error {
	user, exists := r.DB.Users[userID]
	if !exists {
		return errors.ErrUserNotFound
	}
	user.FailedLogins++
	user.UpdatedAt = time.Now()
	return nil
}

// Bloquea al usuario por una duración dada en minutos
func (r *MockUserRepository) LockUser(ctx context.Context, userID string, durationMinutes int) error {
	user, exists := r.DB.Users[userID]
	if !exists {
		return errors.ErrUserNotFound
	}
	if durationMinutes <= 0 {
		until := time.Now().Add(1 * time.Second)
		user.LockedUntil = &until
	} else {
		until := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
		user.LockedUntil = &until
	}
	user.Status = models.UserStatusLocked
	user.UpdatedAt = time.Now()
	return nil
}

// Verifica si el email existe en el tenant dado
func (r *MockUserRepository) EmailExistsInTenant(ctx context.Context, email, tenantID string) (bool, error) {
	key := email + ":" + tenantID
	_, exists := r.DB.EmailTenantIndex[key]
	return exists, nil
}

// Reinicia el contador de intentos fallidos de login
func (r *MockUserRepository) ResetFailedLogins(ctx context.Context, userID string) error {
	user, exists := r.DB.Users[userID]
	if !exists {
		return errors.ErrUserNotFound
	}
	user.FailedLogins = 0
	user.UpdatedAt = time.Now()
	return nil
}

// Actualiza los datos de un usuario existente
func (r *MockUserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	userID := user.ID.String()
	existing, exists := r.DB.Users[userID]
	if !exists {
		return errors.ErrUserNotFound
	}
	// Actualiza los campos principales
	existing.Email = user.Email
	existing.PasswordHash = user.PasswordHash
	existing.FirstName = user.FirstName
	existing.LastName = user.LastName
	existing.TenantID = user.TenantID
	existing.UnitID = user.UnitID
	existing.Role = user.Role
	existing.Status = user.Status
	existing.FailedLogins = user.FailedLogins
	existing.LockedUntil = user.LockedUntil
	existing.UpdatedAt = time.Now()
	return nil
}

// Actualiza el estado de un usuario existente
func (r *MockUserRepository) UpdateUserStatus(ctx context.Context, userID string, status string) error {
	user, exists := r.DB.Users[userID]
	if !exists {
		return errors.ErrUserNotFound
	}
	user.Status = models.UserStatus(status)
	user.UpdatedAt = time.Now()
	return nil
}

// Verifica si el email existe en cualquier tenant
func (r *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	for _, user := range r.DB.Users {
		if user.Email == email {
			return true, nil
		}
	}
	return false, nil
}

// Obtiene un usuario por teléfono (no soportado en el mock, retorna error)
func (r *MockUserRepository) GetUserByPhone(ctx context.Context, phone string) (*models.User, error) {
	return nil, errors.ErrUserNotFound.WithDetails("phone lookup not supported")
}

// Verifica si el teléfono existe (siempre retorna false en el mock)
func (r *MockUserRepository) PhoneExists(ctx context.Context, phone string) (bool, error) {
	return false, nil
}
