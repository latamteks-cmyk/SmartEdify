package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config represents the complete service configuration - aligned with spec requirements
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Security  SecurityConfig  `mapstructure:"security"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	RateLimit RateLimitConfig `mapstructure:"rateLimit"`
	CORS      CORSConfig      `mapstructure:"cors"`
	OpenID    OpenIDConfig    `mapstructure:"openid"`
	
	// Legacy fields for backward compatibility
	Environment string         `json:"environment"`
	HSM         HSMConfig      `json:"hsm"`
	WhatsApp    WhatsAppConfig `json:"whatsapp"`
	Monitoring  MonitoringConfig `json:"monitoring"`
}

// ServerConfig represents server configuration - aligned with spec requirements
type ServerConfig struct {
	Port            int           `mapstructure:"port" default:"8080"`
	Host            string        `mapstructure:"host" default:"0.0.0.0"`
	ReadTimeout     int           `mapstructure:"readTimeout" default:"30"`
	WriteTimeout    int           `mapstructure:"writeTimeout" default:"30"`
	IdleTimeout     int           `mapstructure:"idleTimeout" default:"120"`
	ShutdownTimeout int           `mapstructure:"shutdownTimeout" default:"30"`
	
	// Legacy fields for backward compatibility
	AllowedOrigins string `json:"allowed_origins"`
}

// DatabaseConfig represents database configuration - aligned with spec requirements
type DatabaseConfig struct {
	Host            string `mapstructure:"host" validate:"required"`
	Port            int    `mapstructure:"port" default:"5432"`
	Database        string `mapstructure:"database" validate:"required"`
	Username        string `mapstructure:"username" validate:"required"`
	Password        string `mapstructure:"password" validate:"required"`
	SSLMode         string `mapstructure:"sslMode" default:"require"`
	MaxOpenConns    int    `mapstructure:"maxOpenConns" default:"25"`
	MaxIdleConns    int    `mapstructure:"maxIdleConns" default:"5"`
	ConnMaxLifetime int    `mapstructure:"connMaxLifetime" default:"300"`
	
	// Legacy fields for backward compatibility
	URL  string `json:"url"`
	Name string `json:"name"`
	User string `json:"user"`
}

// RedisConfig represents Redis configuration - aligned with spec requirements
type RedisConfig struct {
	Host         string `mapstructure:"host" validate:"required"`
	Port         int    `mapstructure:"port" default:"6379"`
	Password     string `mapstructure:"password"`
	Database     int    `mapstructure:"database" default:"0"`
	PoolSize     int    `mapstructure:"poolSize" default:"10"`
	MinIdleConns int    `mapstructure:"minIdleConns" default:"5"`
	DialTimeout  int    `mapstructure:"dialTimeout" default:"5"`
	ReadTimeout  int    `mapstructure:"readTimeout" default:"3"`
	WriteTimeout int    `mapstructure:"writeTimeout" default:"3"`
	
	// Legacy fields for backward compatibility
	Address    string `json:"address"`
	DB         int    `json:"db"`
	MaxRetries int    `json:"max_retries"`
}

type HSMConfig struct {
	Endpoint    string `json:"endpoint"`
	Region      string `json:"region"`
	ClusterID   string `json:"cluster_id"`
	KeyID       string `json:"key_id"`
	AccessKeyID string `json:"access_key_id"`
	SecretKey   string `json:"secret_key"`
}

type WhatsAppConfig struct {
	APIEndpoint string `json:"api_endpoint"`
	APIKey      string `json:"api_key"`
	FromNumber  string `json:"from_number"`
	Timeout     int    `json:"timeout"`
}

type JWTConfig struct {
	Issuer           string `json:"issuer"`
	Audience         string `json:"audience"`
	AccessTokenTTL   int    `json:"access_token_ttl"`
	RefreshTokenTTL  int    `json:"refresh_token_ttl"`
	KeyRotationDays  int    `json:"key_rotation_days"`
	KeyVersion       string `json:"key_version"`
}

type SecurityConfig struct {
	EncryptionKey     string `json:"encryption_key"`
	DPoPTTL          int    `json:"dpop_ttl"`
	RateLimitPerIP   int    `json:"rate_limit_per_ip"`
	RateLimitPerUser int    `json:"rate_limit_per_user"`
	MaxLoginAttempts int    `json:"max_login_attempts"`
	BlockDuration    int    `json:"block_duration"`
}

// New configuration types aligned with spec requirements
type LoggingConfig struct {
	Level  string `mapstructure:"level" default:"info"`
	Format string `mapstructure:"format" default:"json"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"true"`
	Port    int    `mapstructure:"port" default:"9090"`
	Path    string `mapstructure:"path" default:"/metrics"`
}

type RateLimitConfig struct {
	LoginAttempts   RateLimit `mapstructure:"loginAttempts"`
	Registration    RateLimit `mapstructure:"registration"`
	TokenValidation RateLimit `mapstructure:"tokenValidation"`
	PasswordReset   RateLimit `mapstructure:"passwordReset"`
	GeneralAPI      RateLimit `mapstructure:"generalAPI"`
}

type RateLimit struct {
	Requests int `mapstructure:"requests"`
	Window   int `mapstructure:"window"` // in seconds
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowedOrigins"`
	AllowedMethods []string `mapstructure:"allowedMethods"`
	AllowedHeaders []string `mapstructure:"allowedHeaders"`
	AllowCredentials bool   `mapstructure:"allowCredentials" default:"true"`
}

type OpenIDConfig struct {
	Issuer                string   `mapstructure:"issuer"`
	AuthorizationEndpoint string   `mapstructure:"authorizationEndpoint"`
	TokenEndpoint         string   `mapstructure:"tokenEndpoint"`
	UserInfoEndpoint      string   `mapstructure:"userInfoEndpoint"`
	JWKSUri              string   `mapstructure:"jwksUri"`
	ScopesSupported      []string `mapstructure:"scopesSupported"`
}

type MonitoringConfig struct {
	PrometheusPort  int    `json:"prometheus_port"`
	JaegerEndpoint  string `json:"jaeger_endpoint"`
	LogLevel        string `json:"log_level"`
	EnableTracing   bool   `json:"enable_tracing"`
	EnableMetrics   bool   `json:"enable_metrics"`
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file is optional, so we don't return an error
	}

	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Server: ServerConfig{
			Port:            getEnvAsInt("SERVER_PORT", 8080),
			Host:            getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:     getEnvAsInt("SERVER_READ_TIMEOUT", 30),
			WriteTimeout:    getEnvAsInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:     getEnvAsInt("SERVER_IDLE_TIMEOUT", 120),
			ShutdownTimeout: getEnvAsInt("SERVER_SHUTDOWN_TIMEOUT", 30),
			AllowedOrigins:  getEnv("ALLOWED_ORIGINS", "*"),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			Database:        getEnv("DB_NAME", "smartedify_auth"),
			Username:        getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			SSLMode:         getEnv("DB_SSL_MODE", "require"),
			MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvAsInt("DB_CONN_MAX_LIFETIME", 300),
			// Legacy fields
			URL:  getEnv("DATABASE_URL", ""),
			Name: getEnv("DB_NAME", "smartedify_auth"),
			User: getEnv("DB_USER", "postgres"),
		},
		Redis: RedisConfig{
			Host:         getEnv("REDIS_HOST", "localhost"),
			Port:         getEnvAsInt("REDIS_PORT", 6379),
			Password:     getEnv("REDIS_PASSWORD", ""),
			Database:     getEnvAsInt("REDIS_DB", 0),
			PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
			MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
			DialTimeout:  getEnvAsInt("REDIS_DIAL_TIMEOUT", 5),
			ReadTimeout:  getEnvAsInt("REDIS_READ_TIMEOUT", 3),
			WriteTimeout: getEnvAsInt("REDIS_WRITE_TIMEOUT", 3),
			// Legacy fields
			Address:    getEnv("REDIS_ADDRESS", "localhost:6379"),
			DB:         getEnvAsInt("REDIS_DB", 0),
			MaxRetries: getEnvAsInt("REDIS_MAX_RETRIES", 3),
		},
		HSM: HSMConfig{
			Endpoint:    getEnv("HSM_ENDPOINT", ""),
			Region:      getEnv("HSM_REGION", "us-east-1"),
			ClusterID:   getEnv("HSM_CLUSTER_ID", ""),
			KeyID:       getEnv("HSM_KEY_ID", ""),
			AccessKeyID: getEnv("HSM_ACCESS_KEY_ID", ""),
			SecretKey:   getEnv("HSM_SECRET_KEY", ""),
		},
		WhatsApp: WhatsAppConfig{
			APIEndpoint: getEnv("WHATSAPP_API_ENDPOINT", ""),
			APIKey:      getEnv("WHATSAPP_API_KEY", ""),
			FromNumber:  getEnv("WHATSAPP_FROM_NUMBER", ""),
			Timeout:     getEnvAsInt("WHATSAPP_TIMEOUT", 30),
		},
		JWT: JWTConfig{
			Issuer:           getEnv("JWT_ISSUER", "smartedify-auth-service"),
			Audience:         getEnv("JWT_AUDIENCE", "smartedify-api"),
			AccessTokenTTL:   getEnvAsInt("JWT_ACCESS_TOKEN_TTL", 3600),
			RefreshTokenTTL:  getEnvAsInt("JWT_REFRESH_TOKEN_TTL", 604800),
			KeyRotationDays:  getEnvAsInt("JWT_KEY_ROTATION_DAYS", 90),
			KeyVersion:       getEnv("JWT_KEY_VERSION", "v1"),
		},
		Security: SecurityConfig{
			EncryptionKey:     getEnv("ENCRYPTION_KEY", ""),
			DPoPTTL:          getEnvAsInt("DPOP_TTL", 300),
			RateLimitPerIP:   getEnvAsInt("RATE_LIMIT_PER_IP", 100),
			RateLimitPerUser: getEnvAsInt("RATE_LIMIT_PER_USER", 50),
			MaxLoginAttempts: getEnvAsInt("MAX_LOGIN_ATTEMPTS", 3),
			BlockDuration:    getEnvAsInt("BLOCK_DURATION", 900),
		},
		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		Metrics: MetricsConfig{
			Enabled: getEnvAsBool("METRICS_ENABLED", true),
			Port:    getEnvAsInt("METRICS_PORT", 9090),
			Path:    getEnv("METRICS_PATH", "/metrics"),
		},
		RateLimit: RateLimitConfig{
			LoginAttempts:   RateLimit{Requests: 5, Window: 60},
			Registration:    RateLimit{Requests: 3, Window: 60},
			TokenValidation: RateLimit{Requests: 100, Window: 60},
			PasswordReset:   RateLimit{Requests: 3, Window: 3600},
			GeneralAPI:      RateLimit{Requests: 1000, Window: 60},
		},
		CORS: CORSConfig{
			AllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "https://smartedify.com"}, ","),
			AllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, ","),
			AllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Authorization"}, ","),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
		OpenID: OpenIDConfig{
			Issuer:                getEnv("OPENID_ISSUER", "https://auth.smartedify.com"),
			AuthorizationEndpoint: getEnv("OPENID_AUTH_ENDPOINT", "/oauth/authorize"),
			TokenEndpoint:         getEnv("OPENID_TOKEN_ENDPOINT", "/oauth/token"),
			UserInfoEndpoint:      getEnv("OPENID_USERINFO_ENDPOINT", "/oauth/userinfo"),
			JWKSUri:              getEnv("OPENID_JWKS_URI", "/.well-known/jwks.json"),
			ScopesSupported:       getEnvAsSlice("OPENID_SCOPES", []string{"openid", "profile", "email"}, ","),
		},
		Monitoring: MonitoringConfig{
			PrometheusPort: getEnvAsInt("PROMETHEUS_PORT", 9090),
			JaegerEndpoint: getEnv("JAEGER_ENDPOINT", "http://localhost:14268/api/traces"),
			LogLevel:       getEnv("LOG_LEVEL", "info"),
			EnableTracing:  getEnvAsBool("ENABLE_TRACING", true),
			EnableMetrics:  getEnvAsBool("ENABLE_METRICS", true),
		},
	}

	// Build database URL if not provided
	if config.Database.URL == "" {
		config.Database.URL = fmt.Sprintf(
			"postgres://%s:%s@%s:%d/%s?sslmode=%s",
			config.Database.Username,
			config.Database.Password,
			config.Database.Host,
			config.Database.Port,
			config.Database.Database,
			config.Database.SSLMode,
		)
	}
	
	// Build Redis address if not provided
	if config.Redis.Address == "" {
		config.Redis.Address = fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, separator string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, separator)
	}
	return defaultValue
}