package conf

import (
	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	RedisUrl    string
)

func init() {
	RedisUrl = "127.0.0.1:6379"
	// TODO - read from env ...
	RedisClient = redis.NewClient(&redis.Options{
		Addr: RedisUrl, // Redis server address
	})
}
