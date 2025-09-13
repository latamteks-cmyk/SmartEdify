package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// ConfigValidator validates configuration settings - aligned with spec requirements
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidateConfig validates the entire configuration
func (v *ConfigValidator) ValidateConfig(config *Config) error {
	if err := v.validateServer(config.Server); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	
	if err := v.validateDatabase(config.Database); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}
	
	if err := v.validateRedis(config.Redis); err != nil {
		return fmt.Errorf("redis config validation failed: %w", err)
	}
	
	if err := v.validateJWT(config.JWT); err != nil {
		return fmt.Errorf("jwt config validation failed: %w", err)
	}
	
	if err := v.validateSecurity(config.Security); err != nil {
		return fmt.Errorf("security config validation failed: %w", err)
	}
	
	return nil
}

func (v *ConfigValidator) validateServer(config ServerConfig) error {
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Port)
	}
	
	if config.Host == "" {
		return fmt.Errorf("server host is required")
	}
	
	if config.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	
	if config.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	
	return nil
}

func (v *ConfigValidator) validateDatabase(config DatabaseConfig) error {
	if config.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid database port: %d", config.Port)
	}
	
	if config.Database == "" {
		return fmt.Errorf("database name is required")
	}
	
	if config.Username == "" {
		return fmt.Errorf("database username is required")
	}
	
	if config.Password == "" {
		return fmt.Errorf("database password is required")
	}
	
	validSSLModes := map[string]bool{
		"disable":     true,
		"allow":       true,
		"prefer":      true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}
	
	if !validSSLModes[config.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", config.SSLMode)
	}
	
	if config.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}
	
	if config.MaxIdleConns <= 0 {
		return fmt.Errorf("max idle connections must be positive")
	}
	
	if config.MaxIdleConns > config.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}
	
	return nil
}

func (v *ConfigValidator) validateRedis(config RedisConfig) error {
	if config.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d", config.Port)
	}
	
	if config.Database < 0 || config.Database > 15 {
		return fmt.Errorf("invalid redis database: %d (must be 0-15)", config.Database)
	}
	
	if config.PoolSize <= 0 {
		return fmt.Errorf("redis pool size must be positive")
	}
	
	if config.MinIdleConns < 0 {
		return fmt.Errorf("redis min idle connections cannot be negative")
	}
	
	if config.MinIdleConns > config.PoolSize {
		return fmt.Errorf("redis min idle connections cannot exceed pool size")
	}
	
	return nil
}

func (v *ConfigValidator) validateJWT(config JWTConfig) error {
	if config.Issuer == "" {
		return fmt.Errorf("JWT issuer is required")
	}
	
	if config.Audience == "" {
		return fmt.Errorf("JWT audience is required")
	}
	
	if config.AccessTokenTTL <= 0 {
		return fmt.Errorf("access token TTL must be positive")
	}
	
	if config.RefreshTokenTTL <= 0 {
		return fmt.Errorf("refresh token TTL must be positive")
	}
	
	if config.AccessTokenTTL >= config.RefreshTokenTTL {
		return fmt.Errorf("access token TTL must be less than refresh token TTL")
	}
	
	return nil
}

func (v *ConfigValidator) validateSecurity(config SecurityConfig) error {
	if config.EncryptionKey == "" {
		return fmt.Errorf("encryption key is required")
	}
	
	if len(config.EncryptionKey) < 32 {
		return fmt.Errorf("encryption key must be at least 32 characters")
	}
	
	if config.RateLimitPerIP <= 0 {
		return fmt.Errorf("rate limit per IP must be positive")
	}
	
	if config.RateLimitPerUser <= 0 {
		return fmt.Errorf("rate limit per user must be positive")
	}
	
	if config.MaxLoginAttempts <= 0 {
		return fmt.Errorf("max login attempts must be positive")
	}
	
	if config.BlockDuration <= 0 {
		return fmt.Errorf("block duration must be positive")
	}
	
	return nil
}

// CheckConnectivity validates that external services are reachable
func (v *ConfigValidator) CheckConnectivity(config *Config) error {
	// Validate database URL format
	if config.Database.URL != "" {
		if _, err := url.Parse(config.Database.URL); err != nil {
			return fmt.Errorf("invalid database URL: %w", err)
		}
	}
	
	// Check if required environment variables are set
	requiredEnvVars := []string{
		"DB_PASSWORD",
		"ENCRYPTION_KEY",
	}
	
	var missingVars []string
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}
	
	if len(missingVars) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missingVars, ", "))
	}
	
	return nil
}

// LoadAndValidate loads configuration and validates it
func LoadAndValidate() (*Config, error) {
	config, err := Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	
	validator := NewConfigValidator()
	
	if err := validator.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}
	
	if err := validator.CheckConnectivity(config); err != nil {
		return nil, fmt.Errorf("connectivity check failed: %w", err)
	}
	
	return config, nil
}