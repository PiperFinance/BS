package conf

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *RedisClientExtended // RedisUrl    string

type RedisClientExtended struct {
	redis.Client
}

func LoadRedis() {
	cl := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", Config.RedisHost, Config.RedisPort),
		DB:   Config.RedisDB,
	})
	RedisClient = &RedisClientExtended{*cl}
	// TODO - Check Redis connection after setup
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
