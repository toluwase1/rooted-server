package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rooted-dating/rooted-server/internal/shared/database"
)

const configCacheTTL = 5 * time.Minute

// DynamicConfig provides access to admin-configurable values stored in PostgreSQL
// with Redis caching. All business-logic values (pricing, algorithm weights, limits)
// live here instead of being hardcoded.
type DynamicConfig struct {
	db    *pgxpool.Pool
	redis *database.SafeRedis
}

func NewDynamicConfig(db *pgxpool.Pool, redis *database.SafeRedis) *DynamicConfig {
	return &DynamicConfig{db: db, redis: redis}
}

func (c *DynamicConfig) hasDB() bool {
	return c.db != nil
}

// getFromDB fetches a raw JSONB value from admin_config. Returns nil if DB unavailable or key not found.
func (c *DynamicConfig) getFromDB(ctx context.Context, key string) []byte {
	if !c.hasDB() {
		return nil
	}
	var val []byte
	err := c.db.QueryRow(ctx, "SELECT value FROM admin_config WHERE key = $1", key).Scan(&val)
	if err != nil {
		return nil
	}
	return val
}

// getRegionalFromDB fetches a regional override. Returns nil if not found.
func (c *DynamicConfig) getRegionalFromDB(ctx context.Context, key, region string) []byte {
	if !c.hasDB() {
		return nil
	}
	var val []byte
	err := c.db.QueryRow(ctx,
		"SELECT value FROM admin_config_regional WHERE config_key = $1 AND region = $2",
		key, region).Scan(&val)
	if err != nil {
		return nil
	}
	return val
}

// GetInt returns an integer config value, checking Redis cache first, then PostgreSQL.
func (c *DynamicConfig) GetInt(ctx context.Context, key string, defaultVal int) int {
	cacheKey := "config:" + key
	if val, err := c.redis.Get(ctx, cacheKey).Int(); err == nil {
		return val
	}
	jsonVal := c.getFromDB(ctx, key)
	if jsonVal == nil {
		return defaultVal
	}
	var result int
	if json.Unmarshal(jsonVal, &result) != nil {
		return defaultVal
	}
	c.redis.Set(ctx, cacheKey, result, configCacheTTL)
	return result
}

// GetFloat returns a float64 config value.
func (c *DynamicConfig) GetFloat(ctx context.Context, key string, defaultVal float64) float64 {
	cacheKey := "config:" + key
	if val, err := c.redis.Get(ctx, cacheKey).Float64(); err == nil {
		return val
	}
	jsonVal := c.getFromDB(ctx, key)
	if jsonVal == nil {
		return defaultVal
	}
	var result float64
	if json.Unmarshal(jsonVal, &result) != nil {
		return defaultVal
	}
	c.redis.Set(ctx, cacheKey, result, configCacheTTL)
	return result
}

// GetString returns a string config value.
func (c *DynamicConfig) GetString(ctx context.Context, key string, defaultVal string) string {
	cacheKey := "config:" + key
	if val, err := c.redis.Get(ctx, cacheKey).Result(); err == nil {
		return val
	}
	jsonVal := c.getFromDB(ctx, key)
	if jsonVal == nil {
		return defaultVal
	}
	var result string
	if json.Unmarshal(jsonVal, &result) != nil {
		return defaultVal
	}
	c.redis.Set(ctx, cacheKey, result, configCacheTTL)
	return result
}

// GetBool returns a boolean config value.
func (c *DynamicConfig) GetBool(ctx context.Context, key string, defaultVal bool) bool {
	cacheKey := "config:" + key
	if val, err := c.redis.Get(ctx, cacheKey).Result(); err == nil {
		return val == "1" || val == "true"
	}
	jsonVal := c.getFromDB(ctx, key)
	if jsonVal == nil {
		return defaultVal
	}
	var result bool
	if json.Unmarshal(jsonVal, &result) != nil {
		return defaultVal
	}

	if result {
		c.redis.Set(ctx, cacheKey, "true", configCacheTTL)
	} else {
		c.redis.Set(ctx, cacheKey, "false", configCacheTTL)
	}
	return result
}

// GetJSON returns a JSON config value unmarshaled into the target.
func (c *DynamicConfig) GetJSON(ctx context.Context, key string, target interface{}) error {
	cacheKey := "config:" + key
	if val, err := c.redis.Get(ctx, cacheKey).Bytes(); err == nil {
		return json.Unmarshal(val, target)
	}
	jsonVal := c.getFromDB(ctx, key)
	if jsonVal == nil {
		return fmt.Errorf("config key %q not found", key)
	}
	c.redis.Set(ctx, cacheKey, string(jsonVal), configCacheTTL)
	return json.Unmarshal(jsonVal, target)
}

// GetIntRegional checks for a region-specific override first, then falls back to global.
func (c *DynamicConfig) GetIntRegional(ctx context.Context, key, region string, defaultVal int) int {
	if region != "" {
		regionalKey := "config:" + key + ":" + region
		if val, err := c.redis.Get(ctx, regionalKey).Int(); err == nil {
			return val
		}
		jsonVal := c.getRegionalFromDB(ctx, key, region)
		if jsonVal != nil {
			var result int
			if json.Unmarshal(jsonVal, &result) == nil {
				c.redis.Set(ctx, regionalKey, result, configCacheTTL)
				return result
			}
		}
	}
	return c.GetInt(ctx, key, defaultVal)
}

// GetFloatRegional checks for a region-specific float override.
func (c *DynamicConfig) GetFloatRegional(ctx context.Context, key, region string, defaultVal float64) float64 {
	if region != "" {
		regionalKey := "config:" + key + ":" + region
		if val, err := c.redis.Get(ctx, regionalKey).Float64(); err == nil {
			return val
		}
		jsonVal := c.getRegionalFromDB(ctx, key, region)
		if jsonVal != nil {
			var result float64
			if json.Unmarshal(jsonVal, &result) == nil {
				c.redis.Set(ctx, regionalKey, result, configCacheTTL)
				return result
			}
		}
	}
	return c.GetFloat(ctx, key, defaultVal)
}

// IsFeatureEnabled checks if a feature flag is enabled for the given context.
// Uses consistent hashing on userID for gradual rollout.
func (c *DynamicConfig) IsFeatureEnabled(ctx context.Context, flag, userID, region, plan string) bool {
	cacheKey := "feature:" + flag

	// Try cache first
	val, err := c.redis.Get(ctx, cacheKey).Bytes()
	if err != nil {
		if !c.hasDB() {
			return false
		}
		// Cache miss → load from DB
		row := c.db.QueryRow(ctx,
			"SELECT enabled, rollout_percent, target_regions, target_plans FROM feature_flags WHERE key = $1",
			flag,
		)

		var enabled bool
		var rolloutPercent int
		var targetRegions, targetPlans []string

		if err := row.Scan(&enabled, &rolloutPercent, &targetRegions, &targetPlans); err != nil {
			return false
		}

		data, _ := json.Marshal(map[string]interface{}{
			"enabled":         enabled,
			"rollout_percent": rolloutPercent,
			"target_regions":  targetRegions,
			"target_plans":    targetPlans,
		})
		c.redis.Set(ctx, cacheKey, data, configCacheTTL)
		val = data
	}

	var ff struct {
		Enabled        bool     `json:"enabled"`
		RolloutPercent int      `json:"rollout_percent"`
		TargetRegions  []string `json:"target_regions"`
		TargetPlans    []string `json:"target_plans"`
	}
	if err := json.Unmarshal(val, &ff); err != nil {
		return false
	}

	if !ff.Enabled {
		return false
	}

	// Check region targeting
	if len(ff.TargetRegions) > 0 && !contains(ff.TargetRegions, region) {
		return false
	}

	// Check plan targeting
	if len(ff.TargetPlans) > 0 && !contains(ff.TargetPlans, plan) {
		return false
	}

	// Gradual rollout — consistent hash on userID
	if ff.RolloutPercent < 100 && ff.RolloutPercent > 0 {
		hash := simpleHash(userID)
		return (hash % 100) < uint32(ff.RolloutPercent)
	}

	return ff.RolloutPercent >= 100
}

// Invalidate removes a config key from Redis cache (called when admin updates a value).
func (c *DynamicConfig) Invalidate(ctx context.Context, key string) {
	c.redis.Del(ctx, "config:"+key)
	// Also delete common regional overrides
	regions := []string{"NG", "KE", "GH", "ZA", "CM", "GB", "US", "CA", "FR", "DE"}
	for _, r := range regions {
		c.redis.Del(ctx, "config:"+key+":"+r)
	}
}

// InvalidateFeature removes a feature flag from Redis cache.
func (c *DynamicConfig) InvalidateFeature(ctx context.Context, flag string) {
	c.redis.Del(ctx, "feature:"+flag)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func simpleHash(s string) uint32 {
	var h uint32
	for _, c := range s {
		h = h*31 + uint32(c)
	}
	return h
}
