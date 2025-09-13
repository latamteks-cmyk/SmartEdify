package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected func(*Config) bool
	}{
		{
			name: "default configuration",
			envVars: map[string]string{
				"DB_PASSWORD":    "testpass",
				"ENCRYPTION_KEY": "test-32-character-encryption-key",
			},
			expected: func(c *Config) bool {
				return c.Server.Port == 8080 &&
					c.Database.Host == "localhost" &&
					c.Redis.Host == "localhost"
			},
		},
		{
			name: "custom configuration",
			envVars: map[string]string{
				"SERVER_PORT":    "9000",
				"DB_HOST":        "custom-db",
				"REDIS_HOST":     "custom-redis",
				"DB_PASSWORD":    "testpass",
				"ENCRYPTION_KEY": "test-32-character-encryption-key",
			},
			expected: func(c *Config) bool {
				return c.Server.Port == 9000 &&
					c.Database.Host == "custom-db" &&
					c.Redis.Host == "custom-redis"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config, err := Load()
			require.NoError(t, err)
			assert.True(t, tt.expected(config))
		})
	}
}

func TestConfigValidator_ValidateConfig(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: &Config{
				Server: ServerConfig{
					Port:        8080,
					Host:        "localhost",
					ReadTimeout: 30,
					WriteTimeout: 30,
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "user",
					Password:     "pass",
					SSLMode:      "require",
					MaxOpenConns: 25,
					MaxIdleConns: 5,
				},
				Redis: RedisConfig{
					Host:         "localhost",
					Port:         6379,
					Database:     0,
					PoolSize:     10,
					MinIdleConns: 5,
				},
				JWT: JWTConfig{
					Issuer:          "test-issuer",
					Audience:        "test-audience",
					AccessTokenTTL:  900,
					RefreshTokenTTL: 604800,
				},
				Security: SecurityConfig{
					EncryptionKey:     "test-32-character-encryption-key",
					RateLimitPerIP:    100,
					RateLimitPerUser:  50,
					MaxLoginAttempts:  5,
					BlockDuration:     1800,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid server port",
			config: &Config{
				Server: ServerConfig{
					Port: -1,
					Host: "localhost",
				},
			},
			wantErr: true,
		},
		{
			name: "missing database host",
			config: &Config{
				Server: ServerConfig{
					Port:        8080,
					Host:        "localhost",
					ReadTimeout: 30,
					WriteTimeout: 30,
				},
				Database: DatabaseConfig{
					Host: "",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid SSL mode",
			config: &Config{
				Server: ServerConfig{
					Port:        8080,
					Host:        "localhost",
					ReadTimeout: 30,
					WriteTimeout: 30,
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "user",
					Password:     "pass",
					SSLMode:      "invalid",
					MaxOpenConns: 25,
					MaxIdleConns: 5,
				},
			},
			wantErr: true,
		},
		{
			name: "short encryption key",
			config: &Config{
				Server: ServerConfig{
					Port:        8080,
					Host:        "localhost",
					ReadTimeout: 30,
					WriteTimeout: 30,
				},
				Database: DatabaseConfig{
					Host:         "localhost",
					Port:         5432,
					Database:     "test",
					Username:     "user",
					Password:     "pass",
					SSLMode:      "require",
					MaxOpenConns: 25,
					MaxIdleConns: 5,
				},
				Redis: RedisConfig{
					Host:         "localhost",
					Port:         6379,
					Database:     0,
					PoolSize:     10,
					MinIdleConns: 5,
				},
				JWT: JWTConfig{
					Issuer:          "test-issuer",
					Audience:        "test-audience",
					AccessTokenTTL:  900,
					RefreshTokenTTL: 604800,
				},
				Security: SecurityConfig{
					EncryptionKey:     "short",
					RateLimitPerIP:    100,
					RateLimitPerUser:  50,
					MaxLoginAttempts:  5,
					BlockDuration:     1800,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadAndValidate(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("ENCRYPTION_KEY", "test-32-character-encryption-key")
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("ENCRYPTION_KEY")
	}()

	config, err := LoadAndValidate()
	require.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 8080, config.Server.Port)
	assert.Equal(t, "localhost", config.Database.Host)
}

func TestLoadAndValidate_MissingRequiredEnvVars(t *testing.T) {
	// Ensure required env vars are not set
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("ENCRYPTION_KEY")

	_, err := LoadAndValidate()
	assert.Error(t, err)
	// The error could be either missing env vars or validation failure
	assert.True(t, 
		err.Error() == "missing required environment variables: DB_PASSWORD, ENCRYPTION_KEY" ||
		err.Error() == "configuration validation failed: database config validation failed: database password is required" ||
		err.Error() == "connectivity check failed: missing required environment variables: DB_PASSWORD, ENCRYPTION_KEY",
		"Expected error about missing required variables or validation failure, got: %s", err.Error())
}

func TestGetEnvHelpers(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		testFunc func()
	}{
		{
			name:     "getEnvAsInt with valid value",
			envVar:   "TEST_INT",
			envValue: "123",
			testFunc: func() {
				result := getEnvAsInt("TEST_INT", 456)
				assert.Equal(t, 123, result)
			},
		},
		{
			name:     "getEnvAsInt with invalid value",
			envVar:   "TEST_INT_INVALID",
			envValue: "invalid",
			testFunc: func() {
				result := getEnvAsInt("TEST_INT_INVALID", 456)
				assert.Equal(t, 456, result)
			},
		},
		{
			name:     "getEnvAsBool with true",
			envVar:   "TEST_BOOL_TRUE",
			envValue: "true",
			testFunc: func() {
				result := getEnvAsBool("TEST_BOOL_TRUE", false)
				assert.True(t, result)
			},
		},
		{
			name:     "getEnvAsBool with false",
			envVar:   "TEST_BOOL_FALSE",
			envValue: "false",
			testFunc: func() {
				result := getEnvAsBool("TEST_BOOL_FALSE", true)
				assert.False(t, result)
			},
		},
		{
			name:     "getEnvAsBool with invalid value",
			envVar:   "TEST_BOOL_INVALID",
			envValue: "invalid",
			testFunc: func() {
				result := getEnvAsBool("TEST_BOOL_INVALID", true)
				assert.True(t, result)
			},
		},
		{
			name:     "getEnvAsSlice with values",
			envVar:   "TEST_SLICE",
			envValue: "a,b,c",
			testFunc: func() {
				result := getEnvAsSlice("TEST_SLICE", []string{"default"}, ",")
				assert.Equal(t, []string{"a", "b", "c"}, result)
			},
		},
		{
			name:     "getEnvAsSlice with default",
			envVar:   "TEST_SLICE_MISSING",
			envValue: "",
			testFunc: func() {
				result := getEnvAsSlice("TEST_SLICE_MISSING", []string{"default"}, ",")
				assert.Equal(t, []string{"default"}, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			}
			tt.testFunc()
		})
	}
}

func TestValidateRedisEdgeCases(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  RedisConfig
		wantErr bool
	}{
		{
			name: "redis database at upper limit",
			config: RedisConfig{
				Host:         "localhost",
				Port:         6379,
				Database:     15,
				PoolSize:     10,
				MinIdleConns: 5,
			},
			wantErr: false,
		},
		{
			name: "redis database above limit",
			config: RedisConfig{
				Host:         "localhost",
				Port:         6379,
				Database:     16,
				PoolSize:     10,
				MinIdleConns: 5,
			},
			wantErr: true,
		},
		{
			name: "min idle conns equals pool size",
			config: RedisConfig{
				Host:         "localhost",
				Port:         6379,
				Database:     0,
				PoolSize:     10,
				MinIdleConns: 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRedis(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateJWTEdgeCases(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  JWTConfig
		wantErr bool
	}{
		{
			name: "access token TTL equals refresh token TTL",
			config: JWTConfig{
				Issuer:          "test",
				Audience:        "test",
				AccessTokenTTL:  900,
				RefreshTokenTTL: 900,
			},
			wantErr: true,
		},
		{
			name: "access token TTL greater than refresh token TTL",
			config: JWTConfig{
				Issuer:          "test",
				Audience:        "test",
				AccessTokenTTL:  1000,
				RefreshTokenTTL: 900,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateJWT(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckConnectivity(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  *Config
		envVars map[string]string
		wantErr bool
	}{
		{
			name: "valid connectivity with all env vars",
			config: &Config{
				Database: DatabaseConfig{
					URL: "postgres://user:pass@localhost:5432/db",
				},
			},
			envVars: map[string]string{
				"DB_PASSWORD":    "testpass",
				"ENCRYPTION_KEY": "test-32-character-encryption-key",
			},
			wantErr: false,
		},
		{
			name: "invalid database URL",
			config: &Config{
				Database: DatabaseConfig{
					URL: "://invalid-url",
				},
			},
			envVars: map[string]string{
				"DB_PASSWORD":    "testpass",
				"ENCRYPTION_KEY": "test-32-character-encryption-key",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			err := validator.CheckConnectivity(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDatabaseEdgeCases(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
	}{
		{
			name: "max idle conns greater than max open conns",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "test",
				Username:     "user",
				Password:     "pass",
				SSLMode:      "require",
				MaxOpenConns: 5,
				MaxIdleConns: 10,
			},
			wantErr: true,
		},
		{
			name: "zero max open conns",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "test",
				Username:     "user",
				Password:     "pass",
				SSLMode:      "require",
				MaxOpenConns: 0,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
		{
			name: "zero max idle conns",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "test",
				Username:     "user",
				Password:     "pass",
				SSLMode:      "require",
				MaxOpenConns: 25,
				MaxIdleConns: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid port range",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         70000,
				Database:     "test",
				Username:     "user",
				Password:     "pass",
				SSLMode:      "require",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateDatabase(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateServerEdgeCases(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name    string
		config  ServerConfig
		wantErr bool
	}{
		{
			name: "port at upper limit",
			config: ServerConfig{
				Port:         65535,
				Host:         "localhost",
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			wantErr: false,
		},
		{
			name: "port above upper limit",
			config: ServerConfig{
				Port:         65536,
				Host:         "localhost",
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			wantErr: true,
		},
		{
			name: "zero read timeout",
			config: ServerConfig{
				Port:         8080,
				Host:         "localhost",
				ReadTimeout:  0,
				WriteTimeout: 30,
			},
			wantErr: true,
		},
		{
			name: "zero write timeout",
			config: ServerConfig{
				Port:         8080,
				Host:         "localhost",
				ReadTimeout:  30,
				WriteTimeout: 0,
			},
			wantErr: true,
		},
		{
			name: "empty host",
			config: ServerConfig{
				Port:         8080,
				Host:         "",
				ReadTimeout:  30,
				WriteTimeout: 30,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateServer(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadWithDotEnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := `SERVER_PORT=9999
DB_HOST=test-db
REDIS_HOST=test-redis
DB_PASSWORD=testpass
ENCRYPTION_KEY=test-32-character-encryption-key`

	err := os.WriteFile(".env", []byte(envContent), 0644)
	require.NoError(t, err)
	defer os.Remove(".env")

	config, err := Load()
	require.NoError(t, err)
	
	assert.Equal(t, 9999, config.Server.Port)
	assert.Equal(t, "test-db", config.Database.Host)
	assert.Equal(t, "test-redis", config.Redis.Host)
}

func TestConfigURLBuilding(t *testing.T) {
	// Set required environment variables
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("ENCRYPTION_KEY", "test-32-character-encryption-key")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_SSL_MODE", "disable")
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("ENCRYPTION_KEY")
		os.Unsetenv("DB_USER")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_SSL_MODE")
	}()

	config, err := Load()
	require.NoError(t, err)
	
	expectedURL := "postgres://testuser:testpass@testhost:5433/testdb?sslmode=disable"
	assert.Equal(t, expectedURL, config.Database.URL)
	
	expectedRedisAddr := "localhost:6379"
	assert.Equal(t, expectedRedisAddr, config.Redis.Address)
}

func TestGetEnvMissingValues(t *testing.T) {
	// Test getEnv with missing environment variable
	result := getEnv("MISSING_ENV_VAR", "default_value")
	assert.Equal(t, "default_value", result)
	
	// Test getEnvAsInt with missing environment variable
	intResult := getEnvAsInt("MISSING_INT_VAR", 42)
	assert.Equal(t, 42, intResult)
	
	// Test getEnvAsBool with missing environment variable
	boolResult := getEnvAsBool("MISSING_BOOL_VAR", true)
	assert.True(t, boolResult)
	
	// Test getEnvAsSlice with missing environment variable
	sliceResult := getEnvAsSlice("MISSING_SLICE_VAR", []string{"default1", "default2"}, ",")
	assert.Equal(t, []string{"default1", "default2"}, sliceResult)
}

func TestValidateAllSSLModes(t *testing.T) {
	validator := NewConfigValidator()
	
	validSSLModes := []string{"disable", "allow", "prefer", "require", "verify-ca", "verify-full"}
	
	for _, sslMode := range validSSLModes {
		t.Run("ssl_mode_"+sslMode, func(t *testing.T) {
			config := DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "test",
				Username:     "user",
				Password:     "pass",
				SSLMode:      sslMode,
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			}
			
			err := validator.validateDatabase(config)
			assert.NoError(t, err)
		})
	}
}

func TestValidateRedisPortEdgeCases(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name    string
		port    int
		wantErr bool
	}{
		{"port_1", 1, false},
		{"port_65535", 65535, false},
		{"port_0", 0, true},
		{"port_negative", -1, true},
		{"port_too_high", 65536, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RedisConfig{
				Host:         "localhost",
				Port:         tt.port,
				Database:     0,
				PoolSize:     10,
				MinIdleConns: 5,
			}
			
			err := validator.validateRedis(config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigWithExistingDatabaseURL(t *testing.T) {
	// Set environment variables including DATABASE_URL
	os.Setenv("DATABASE_URL", "postgres://existing:pass@existing:5432/existing?sslmode=require")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("ENCRYPTION_KEY", "test-32-character-encryption-key")
	defer func() {
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("ENCRYPTION_KEY")
	}()

	config, err := Load()
	require.NoError(t, err)
	
	// Should use the existing DATABASE_URL
	assert.Equal(t, "postgres://existing:pass@existing:5432/existing?sslmode=require", config.Database.URL)
}

func TestConfigWithExistingRedisAddress(t *testing.T) {
	// Set environment variables including REDIS_ADDRESS
	os.Setenv("REDIS_ADDRESS", "existing-redis:6380")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("ENCRYPTION_KEY", "test-32-character-encryption-key")
	defer func() {
		os.Unsetenv("REDIS_ADDRESS")
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("ENCRYPTION_KEY")
	}()

	config, err := Load()
	require.NoError(t, err)
	
	// Should use the existing REDIS_ADDRESS
	assert.Equal(t, "existing-redis:6380", config.Redis.Address)
}

func TestNewConfigValidator(t *testing.T) {
	validator := NewConfigValidator()
	assert.NotNil(t, validator)
}

func TestValidateSecurityEdgeCases(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name    string
		config  SecurityConfig
		wantErr bool
	}{
		{
			name: "encryption key exactly 32 chars",
			config: SecurityConfig{
				EncryptionKey:     "12345678901234567890123456789012",
				RateLimitPerIP:    100,
				RateLimitPerUser:  50,
				MaxLoginAttempts:  5,
				BlockDuration:     1800,
			},
			wantErr: false,
		},
		{
			name: "encryption key 31 chars",
			config: SecurityConfig{
				EncryptionKey:     "1234567890123456789012345678901",
				RateLimitPerIP:    100,
				RateLimitPerUser:  50,
				MaxLoginAttempts:  5,
				BlockDuration:     1800,
			},
			wantErr: true,
		},
		{
			name: "zero rate limit per IP",
			config: SecurityConfig{
				EncryptionKey:     "test-32-character-encryption-key",
				RateLimitPerIP:    0,
				RateLimitPerUser:  50,
				MaxLoginAttempts:  5,
				BlockDuration:     1800,
			},
			wantErr: true,
		},
		{
			name: "zero rate limit per user",
			config: SecurityConfig{
				EncryptionKey:     "test-32-character-encryption-key",
				RateLimitPerIP:    100,
				RateLimitPerUser:  0,
				MaxLoginAttempts:  5,
				BlockDuration:     1800,
			},
			wantErr: true,
		},
		{
			name: "zero max login attempts",
			config: SecurityConfig{
				EncryptionKey:     "test-32-character-encryption-key",
				RateLimitPerIP:    100,
				RateLimitPerUser:  50,
				MaxLoginAttempts:  0,
				BlockDuration:     1800,
			},
			wantErr: true,
		},
		{
			name: "zero block duration",
			config: SecurityConfig{
				EncryptionKey:     "test-32-character-encryption-key",
				RateLimitPerIP:    100,
				RateLimitPerUser:  50,
				MaxLoginAttempts:  5,
				BlockDuration:     0,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateSecurity(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateJWTMissingFields(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name    string
		config  JWTConfig
		wantErr bool
	}{
		{
			name: "missing issuer",
			config: JWTConfig{
				Issuer:          "",
				Audience:        "test",
				AccessTokenTTL:  900,
				RefreshTokenTTL: 604800,
			},
			wantErr: true,
		},
		{
			name: "missing audience",
			config: JWTConfig{
				Issuer:          "test",
				Audience:        "",
				AccessTokenTTL:  900,
				RefreshTokenTTL: 604800,
			},
			wantErr: true,
		},
		{
			name: "zero access token TTL",
			config: JWTConfig{
				Issuer:          "test",
				Audience:        "test",
				AccessTokenTTL:  0,
				RefreshTokenTTL: 604800,
			},
			wantErr: true,
		},
		{
			name: "zero refresh token TTL",
			config: JWTConfig{
				Issuer:          "test",
				Audience:        "test",
				AccessTokenTTL:  900,
				RefreshTokenTTL: 0,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateJWT(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDatabaseMissingFields(t *testing.T) {
	validator := NewConfigValidator()
	
	tests := []struct {
		name    string
		config  DatabaseConfig
		wantErr bool
	}{
		{
			name: "missing database name",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "",
				Username:     "user",
				Password:     "pass",
				SSLMode:      "require",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: DatabaseConfig{
				Host:         "localhost",
				Port:         5432,
				Database:     "test",
				Username:     "",
				Password:     "pass",
				SSLMode:      "require",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateDatabase(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRedisNegativeMinIdleConns(t *testing.T) {
	validator := NewConfigValidator()
	
	config := RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Database:     0,
		PoolSize:     10,
		MinIdleConns: -1,
	}
	
	err := validator.validateRedis(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis min idle connections cannot be negative")
}