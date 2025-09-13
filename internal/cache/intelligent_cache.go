package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/smartedify/auth-service/internal/errors"
)

type IntelligentCache struct {
    client *redis.Client
    stats  *CacheStats
}

type CacheStats struct {
    Hits   int64
    Misses int64
    Evictions int64
}

func NewIntelligentCache(redisURL string) *IntelligentCache {
    rdb := redis.NewClient(&redis.Options{
        Addr: redisURL,
        DB:   0,
    })
    
    return &IntelligentCache{
        client: rdb,
        stats:  &CacheStats{},
    }
}

func (ic *IntelligentCache) GetWithTTLAdaptive(ctx context.Context, key string, dest interface{}) error {
    val, err := ic.client.Get(ctx, key).Result()
    if err == redis.Nil {
        ic.stats.Misses++
        return errors.ErrCacheMiss
    } else if err != nil {
        return err
    }
    
    ic.stats.Hits++
    return json.Unmarshal([]byte(val), dest)
}

func (ic *IntelligentCache) SetWithAdaptiveTTL(ctx context.Context, key string, value interface{}, baseExpiration time.Duration) error {
    // Adaptive TTL based on access patterns
    hitRatio := float64(ic.stats.Hits) / float64(ic.stats.Hits + ic.stats.Misses)
    
    var ttl time.Duration
    if hitRatio > 0.8 {
        ttl = baseExpiration * 2 // High hit ratio, extend TTL
    } else if hitRatio < 0.3 {
        ttl = baseExpiration / 2 // Low hit ratio, reduce TTL
    } else {
        ttl = baseExpiration
    }
    
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    return ic.client.Set(ctx, key, data, ttl).Err()
}