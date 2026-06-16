package infra

import (
	"context"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/redis/go-redis/v9"
)

type CacheWrapper struct {
	*redis.Client
	cache	*cache.Cache
}

type PatternDeleter interface {
	DelByPattern(ctx context.Context, pattern string) int
}

func NewCacheWrapper(defaultExpiration, cleanupInterval time.Duration) *CacheWrapper {
	return &CacheWrapper{
		Client:	&redis.Client{},
		cache:	cache.New(defaultExpiration, cleanupInterval),
	}
}

func (c *CacheWrapper) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx)
	val, found := c.cache.Get(key)
	if !found {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	switch v := val.(type) {
	case string:
		cmd.SetVal(v)
	case []byte:
		cmd.SetVal(string(v))
	default:
		cmd.SetErr(redis.Nil)
	}
	return cmd
}

func (c *CacheWrapper) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx)
	if expiration == 0 {
		expiration = cache.NoExpiration
	}
	c.cache.Set(key, value, expiration)
	cmd.SetVal("OK")
	return cmd
}

func (c *CacheWrapper) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx)
	if expiration == 0 {
		expiration = cache.NoExpiration
	}
	err := c.cache.Add(key, value, expiration)
	cmd.SetVal(err == nil)
	return cmd
}

func (c *CacheWrapper) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	var count int64
	for _, key := range keys {
		if _, found := c.cache.Get(key); found {
			c.cache.Delete(key)
			count++
		}
	}
	cmd.SetVal(count)
	return cmd
}

func (c *CacheWrapper) ItemCount() int {
	return c.cache.ItemCount()
}

func (c *CacheWrapper) DelByPattern(ctx context.Context, pattern string) int {
	items := c.cache.Items()
	var keys []string
	for k := range items {
		if strings.HasPrefix(k, pattern) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return 0
	}
	cmd := c.Del(ctx, keys...)
	if cmd.Err() != nil {
		return 0
	}
	return int(cmd.Val())
}
