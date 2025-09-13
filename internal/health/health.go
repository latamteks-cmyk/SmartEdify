package health

import (
	"context"
	"database/sql"
	"time"

	"github.com/redis/go-redis/v9"
)

// HealthChecker interface defines health check operations - aligned with spec requirements
type HealthChecker interface {
	CheckDatabase(ctx context.Context) error
	CheckRedis(ctx context.Context) error
	CheckKeyStore(ctx context.Context) error
}

type healthChecker struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB, redis *redis.Client) HealthChecker {
	return &healthChecker{
		db:    db,
		redis: redis,
	}
}

func (h *healthChecker) CheckDatabase(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	return h.db.PingContext(ctx)
}

func (h *healthChecker) CheckRedis(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	return h.redis.Ping(ctx).Err()
}

func (h *healthChecker) CheckKeyStore(ctx context.Context) error {
	// For now, just return nil as we're using mock HSM
	// In production, this would check HSM connectivity
	return nil
}

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status      string            `json:"status"`
	Timestamp   time.Time         `json:"timestamp"`
	Service     string            `json:"service"`
	Version     string            `json:"version"`
	Checks      map[string]string `json:"checks,omitempty"`
	Environment string            `json:"environment,omitempty"`
}