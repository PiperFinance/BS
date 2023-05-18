package conf

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-redis/redis/v8"
)

const RedisUrl = "redis://redis:6379/2"

var RedisClient *RedisClientExtended // RedisUrl    string

type RedisClientExtended struct {
	redis.Client
}

func init() {
	// RedisUrl = "redis://redis:6379/2"
	// TODO - read from env ...
	opts, err := redis.ParseURL(RedisUrl)
	if err != nil {
		log.Fatalf("Redis: %s", err)
	}
	cl := redis.NewClient(opts)
	RedisClient = &RedisClientExtended{*cl}
}

func (r *RedisClientExtended) GetOrSet(context context.Context, key string, value string) (string, error) {
	return r.GetOrSetTTL(context, key, value, redis.KeepTTL)
}

func (r *RedisClientExtended) GetOrSetTTL(
	context context.Context, key string, value string, ttl time.Duration,
) (string, error) {
	if res := r.Get(context, key); res.Err() != nil {
		if res.Err() == redis.Nil {
			if res := r.Set(context, key, value, ttl); res.Err() != nil {
				return "", res.Err()
			}
		} else {
			return "", res.Err()
		}
	} else {
		value = res.Val()
	}
	return value, nil
}
