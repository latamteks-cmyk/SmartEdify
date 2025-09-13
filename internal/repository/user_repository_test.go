package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/smartedify/auth-service/internal/errors"
	"github.com/smartedify/auth-service/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
	mock "github.com/smartedify/auth-service/internal/repository"
)

func TestUserRepository_CreateUser(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("create user successfully", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		assert.NoError(t, err)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})

	t.Run("create user with duplicate email in same tenant should fail", func(t *testing.T) {
		user1 := &models.User{
			ID:           uuid.New(),
			Email:        "duplicate@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user1)
		assert.NoError(t, err)

		user2 := &models.User{
			ID:           uuid.New(),
			Email:        "duplicate@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "Jane",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-789",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err = repo.CreateUser(context.Background(), user2)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrEmailAlreadyExists, err)
	})

	t.Run("create user with same email in different tenant should succeed", func(t *testing.T) {
		user1 := &models.User{
			ID:           uuid.New(),
			Email:        "same@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user1)
		assert.NoError(t, err)

		user2 := &models.User{
			ID:           uuid.New(),
			Email:        "same@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "Jane",
			LastName:     "Doe",
			TenantID:     "tenant-456",
			UnitID:       "unit-789",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err = repo.CreateUser(context.Background(), user2)
		assert.NoError(t, err)
	})
}

func TestUserRepository_GetUserByEmailAndTenant(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("get existing user", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Get the user
		foundUser, err := repo.GetUserByEmailAndTenant(context.Background(), "test@example.com", "tenant-123")
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.TenantID, foundUser.TenantID)
		assert.Equal(t, user.FirstName, foundUser.FirstName)
	})

	t.Run("get non-existent user should return error", func(t *testing.T) {
		user, err := repo.GetUserByEmailAndTenant(context.Background(), "nonexistent@example.com", "tenant-123")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, user)
	})

	t.Run("get user with wrong tenant should return error", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "tenant-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Try to get with wrong tenant
		foundUser, err := repo.GetUserByEmailAndTenant(context.Background(), "tenant-test@example.com", "wrong-tenant")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, foundUser)
	})
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("update last login successfully", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "login-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Update last login
		ipAddress := "192.168.1.1"
		userAgent := "Mozilla/5.0 Test Browser"
		err = repo.UpdateLastLogin(context.Background(), user.ID.String(), ipAddress, userAgent)
		assert.NoError(t, err)

		// Verify the update
		updatedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser.LastLoginAt)
		assert.Equal(t, ipAddress, updatedUser.LastLoginIP)
	})

	t.Run("update last login for non-existent user should return error", func(t *testing.T) {
		err := repo.UpdateLastLogin(context.Background(), uuid.New().String(), "192.168.1.1", "Test Browser")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

func TestUserRepository_IncrementFailedLogins(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("increment failed logins", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "failed-login-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
			FailedLogins: 0,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Increment failed logins
		err = repo.IncrementFailedLogins(context.Background(), user.ID.String())
		assert.NoError(t, err)

		// Verify the increment
		updatedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, 1, updatedUser.FailedLogins)

		// Increment again
		err = repo.IncrementFailedLogins(context.Background(), user.ID.String())
		assert.NoError(t, err)

		updatedUser, err = repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, 2, updatedUser.FailedLogins)
	})

	t.Run("increment failed logins for non-existent user should return error", func(t *testing.T) {
		err := repo.IncrementFailedLogins(context.Background(), uuid.New().String())
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

func TestUserRepository_LockUser(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("lock user for specified duration", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "lock-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Lock the user
		lockDurationMinutes := 30
		err = repo.LockUser(context.Background(), user.ID.String(), lockDurationMinutes)
		assert.NoError(t, err)

		// Verify the user is locked
		lockedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.UserStatusLocked, lockedUser.Status)
		assert.NotNil(t, lockedUser.LockedUntil)
		assert.True(t, lockedUser.LockedUntil.After(time.Now()))
	})

	t.Run("lock non-existent user should return error", func(t *testing.T) {
		err := repo.LockUser(context.Background(), uuid.New().String(), 30)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

func TestUserRepository_EmailExistsInTenant(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("email exists in tenant", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "existing@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Check if email exists in tenant
		exists, err := repo.EmailExistsInTenant(context.Background(), "existing@example.com", "tenant-123")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("email does not exist in tenant", func(t *testing.T) {
		exists, err := repo.EmailExistsInTenant(context.Background(), "nonexistent@example.com", "tenant-123")
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("email exists in different tenant", func(t *testing.T) {
		// Create a user in tenant-123
		user := &models.User{
			ID:           uuid.New(),
			Email:        "cross-tenant@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Check if email exists in different tenant
		exists, err := repo.EmailExistsInTenant(context.Background(), "cross-tenant@example.com", "tenant-456")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// Helper function to create a test user
func createTestUser() *models.User {
	return &models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashed_password_here",
		FirstName:    "John",
		LastName:     "Doe",
		TenantID:     "tenant-123",
		UnitID:       "unit-456",
		Role:         "user",
		Status:       models.UserStatusActive,
		FailedLogins: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestUserRepository_ResetFailedLogins(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("reset failed logins successfully", func(t *testing.T) {
		// Create a user with failed logins
		user := &models.User{
			ID:           uuid.New(),
			Email:        "reset-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
			FailedLogins: 3,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Reset failed logins
		err = repo.ResetFailedLogins(context.Background(), user.ID.String())
		assert.NoError(t, err)

		// Verify the reset
		updatedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, 0, updatedUser.FailedLogins)
		assert.Nil(t, updatedUser.LockedUntil)
	})

	t.Run("reset failed logins for non-existent user should return error", func(t *testing.T) {
		err := repo.ResetFailedLogins(context.Background(), uuid.New().String())
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

func TestUserRepository_UpdateUser(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("update user successfully", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "update-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Update the user
		user.FirstName = "Jane"
		user.LastName = "Smith"
		user.Role = "admin"

		err = repo.UpdateUser(context.Background(), user)
		assert.NoError(t, err)

		// Verify the update
		updatedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, "Jane", updatedUser.FirstName)
		assert.Equal(t, "Smith", updatedUser.LastName)
		assert.Equal(t, "admin", updatedUser.Role)
	})

	t.Run("update non-existent user should return error", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.New(),
			Email:        "nonexistent@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.UpdateUser(context.Background(), user)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

func TestUserRepository_UpdateUserStatus(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("update user status successfully", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "status-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Update status
		err = repo.UpdateUserStatus(context.Background(), user.ID.String(), "suspended")
		assert.NoError(t, err)

		// Verify the update
		updatedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.UserStatusSuspended, updatedUser.Status)
	})

	t.Run("update status for non-existent user should return error", func(t *testing.T) {
		err := repo.UpdateUserStatus(context.Background(), uuid.New().String(), "suspended")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("get user by email successfully", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "email-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Get user by email
		foundUser, err := repo.GetUserByEmail(context.Background(), "email-test@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user.Email, foundUser.Email)
		assert.Equal(t, user.FirstName, foundUser.FirstName)
	})

	t.Run("get non-existent user by email should return error", func(t *testing.T) {
		user, err := repo.GetUserByEmail(context.Background(), "nonexistent@example.com")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, user)
	})
}

func TestUserRepository_EmailExists(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("email exists", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "exists-test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Check if email exists
		exists, err := repo.EmailExists(context.Background(), "exists-test@example.com")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("email does not exist", func(t *testing.T) {
		exists, err := repo.EmailExists(context.Background(), "does-not-exist@example.com")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// Integration test setup helper (would be used in real tests)
func setupTestDB(t *testing.T) *sql.DB {
	// This would set up a test database connection
	// For now, return nil as we're not running real database tests
	return nil
}

// Integration test cleanup helper (would be used in real tests)
func cleanupTestDB(t *testing.T, db *sql.DB) {
	// This would clean up test data and close the database connection
}

func TestNewUserRepository(t *testing.T) {
	t.Run("create new user repository", func(t *testing.T) {
		repo := mock.NewMockUserRepository()
		assert.NotNil(t, repo)
	})
}

func TestUserRepository_GetUserByPhone(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("get user by phone should return not supported error", func(t *testing.T) {
		user, err := repo.GetUserByPhone(context.Background(), "+1234567890")
		assert.Error(t, err)
		assert.Nil(t, user)
		// The actual implementation returns ErrUserNotFound with details about phone not being supported
		apiErr, ok := err.(*errors.APIError)
		assert.True(t, ok)
		assert.Equal(t, "USER_NOT_FOUND", apiErr.Code)
		assert.Contains(t, apiErr.Details, "phone lookup not supported")
	})
}

func TestUserRepository_PhoneExists(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("phone exists should always return false", func(t *testing.T) {
		exists, err := repo.PhoneExists(context.Background(), "+1234567890")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// Test interface compliance

// Test model validation
func TestUserModel(t *testing.T) {
	t.Run("create user model with all fields", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.New(),
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
			FailedLogins: 0,
		}

		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "tenant-123", user.TenantID)
		assert.Equal(t, "unit-456", user.UnitID)
		assert.Equal(t, "user", user.Role)
		assert.Equal(t, models.UserStatusActive, user.Status)
		assert.Equal(t, 0, user.FailedLogins)
	})

	t.Run("test user status constants", func(t *testing.T) {
		assert.Equal(t, models.UserStatus("active"), models.UserStatusActive)
		assert.Equal(t, models.UserStatus("inactive"), models.UserStatusInactive)
		assert.Equal(t, models.UserStatus("suspended"), models.UserStatusSuspended)
		assert.Equal(t, models.UserStatus("locked"), models.UserStatusLocked)
	})
}

// Test error handling
func TestRepositoryErrorHandling(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("methods with non-existent user return appropriate errors", func(t *testing.T) {
		nonExistentID := uuid.New().String()

		// Test GetUserByID
		user, err := repo.GetUserByID(context.Background(), nonExistentID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
		assert.Nil(t, user)

		// Test UpdateUser
		testUser := &models.User{ID: uuid.New()}
		err = repo.UpdateUser(context.Background(), testUser)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)

		// Test UpdateUserStatus
		err = repo.UpdateUserStatus(context.Background(), nonExistentID, "active")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)

		// Test UpdateLastLogin
		err = repo.UpdateLastLogin(context.Background(), nonExistentID, "127.0.0.1", "Test Browser")
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)

		// Test IncrementFailedLogins
		err = repo.IncrementFailedLogins(context.Background(), nonExistentID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)

		// Test ResetFailedLogins
		err = repo.ResetFailedLogins(context.Background(), nonExistentID)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)

		// Test LockUser
		err = repo.LockUser(context.Background(), nonExistentID, 30)
		assert.Error(t, err)
		assert.Equal(t, errors.ErrUserNotFound, err)
	})
}

// Test edge cases
func TestRepositoryEdgeCases(t *testing.T) {
	repo := mock.NewMockUserRepository()

	t.Run("create user with empty ID should work", func(t *testing.T) {
		user := &models.User{
			ID:           uuid.Nil, // Empty UUID
			Email:        "empty-id@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		assert.NoError(t, err)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})

	t.Run("lock user with zero duration", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "zero-lock@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Lock with zero duration
		err = repo.LockUser(context.Background(), user.ID.String(), 0)
		assert.NoError(t, err)

		// Verify the user is locked
		lockedUser, err := repo.GetUserByID(context.Background(), user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, models.UserStatusLocked, lockedUser.Status)
		assert.NotNil(t, lockedUser.LockedUntil)
	})

	t.Run("update user with same data should work", func(t *testing.T) {
		// Create a user first
		user := &models.User{
			ID:           uuid.New(),
			Email:        "same-data@example.com",
			PasswordHash: "hashed_password",
			FirstName:    "John",
			LastName:     "Doe",
			TenantID:     "tenant-123",
			UnitID:       "unit-456",
			Role:         "user",
			Status:       models.UserStatusActive,
		}

		err := repo.CreateUser(context.Background(), user)
		require.NoError(t, err)

		// Update with same data
		err = repo.UpdateUser(context.Background(), user)
		assert.NoError(t, err)
	})
}
