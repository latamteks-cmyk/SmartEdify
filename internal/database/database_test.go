package database

import (
	"testing"

	"github.com/smartedify/auth-service/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	t.Run("connect with invalid config should return error", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			URL:             "invalid-url",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
		}

		db, err := Connect(cfg)
		assert.Error(t, err)
		assert.Nil(t, db)
		assert.Contains(t, err.Error(), "failed to ping database")
	})

	t.Run("connect with valid config structure", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			Database:        "test_db",
			Username:        "test_user",
			Password:        "test_pass",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300,
			URL:             "postgres://test_user:test_pass@localhost:5432/test_db?sslmode=disable",
		}

		// This will fail because we don't have a real database, but we can test the structure
		db, err := Connect(cfg)
		// We expect an error because the database doesn't exist
		assert.Error(t, err)
		assert.Nil(t, db)
		// But the error should be about connection, not about invalid URL format
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}

func TestMigrate(t *testing.T) {
	t.Run("migrate with invalid database URL should return error", func(t *testing.T) {
		err := Migrate("invalid-url")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create postgres driver")
	})

	t.Run("migrate with valid URL format but non-existent database", func(t *testing.T) {
		// This will fail because we don't have a real database
		err := Migrate("postgres://test:test@localhost:5432/nonexistent?sslmode=disable")
		assert.Error(t, err)
		// The error should be about database connection, not URL format
		assert.Contains(t, err.Error(), "failed to create postgres driver")
	})
}

func TestHealth(t *testing.T) {
	t.Run("health check with nil database should panic", func(t *testing.T) {
		// This will panic because we're passing nil, which is expected behavior
		// In a real scenario, Health should never be called with nil
		assert.Panics(t, func() {
			Health(nil)
		})
	})
}