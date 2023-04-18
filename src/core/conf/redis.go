package conf

import (
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

var (
	RedisClient *redis.Client
	RedisUrl    string
)

func init() {
	RedisUrl = "redis://localhost:6379"
	// TODO - read from env ...
	opts, err := redis.ParseURL(RedisUrl)
	if err != nil {
		log.Fatalf("Redis: %s", err)
	}
	RedisClient = redis.NewClient(opts)
}
