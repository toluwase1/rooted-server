package database

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// SafeRedis wraps redis.Client and handles nil gracefully.
// All operations return zero values / errors when Redis is not available,
// so the app functions without caching.
type SafeRedis struct {
	client *redis.Client
}

func NewSafeRedis(client *redis.Client) *SafeRedis {
	return &SafeRedis{client: client}
}

// Available returns true if Redis is connected.
func (r *SafeRedis) Available() bool {
	return r.client != nil
}

// Inner returns the underlying client (for packages that need it directly).
func (r *SafeRedis) Inner() *redis.Client {
	return r.client
}

func (r *SafeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	if r.client == nil {
		cmd := redis.NewStringCmd(ctx)
		cmd.SetErr(redis.Nil)
		return cmd
	}
	return r.client.Get(ctx, key)
}

func (r *SafeRedis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	if r.client == nil {
		return redis.NewStatusCmd(ctx)
	}
	return r.client.Set(ctx, key, value, expiration)
}

func (r *SafeRedis) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	if r.client == nil {
		return redis.NewIntCmd(ctx)
	}
	return r.client.Del(ctx, keys...)
}

func (r *SafeRedis) Incr(ctx context.Context, key string) *redis.IntCmd {
	if r.client == nil {
		return redis.NewIntCmd(ctx)
	}
	return r.client.Incr(ctx, key)
}

func (r *SafeRedis) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	if r.client == nil {
		return redis.NewBoolCmd(ctx)
	}
	return r.client.Expire(ctx, key, expiration)
}

func (r *SafeRedis) TTL(ctx context.Context, key string) *redis.DurationCmd {
	if r.client == nil {
		cmd := redis.NewDurationCmd(ctx, time.Duration(0))
		cmd.SetErr(redis.Nil)
		return cmd
	}
	return r.client.TTL(ctx, key)
}

func (r *SafeRedis) SAdd(ctx context.Context, key string, members ...interface{}) *redis.IntCmd {
	if r.client == nil {
		return redis.NewIntCmd(ctx)
	}
	return r.client.SAdd(ctx, key, members...)
}

func (r *SafeRedis) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	if r.client == nil {
		cmd := redis.NewIntCmd(ctx)
		return cmd
	}
	return r.client.Exists(ctx, keys...)
}

func (r *SafeRedis) RPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	if r.client == nil {
		return redis.NewIntCmd(ctx)
	}
	return r.client.RPush(ctx, key, values...)
}

func (r *SafeRedis) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	if r.client == nil {
		return redis.NewScanCmd(ctx, nil)
	}
	return r.client.Scan(ctx, cursor, match, count)
}

func (r *SafeRedis) Close() error {
	if r.client == nil {
		return nil
	}
	return r.client.Close()
}
