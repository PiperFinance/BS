package conf

import (
	"github.com/go-redis/redis/v8"
)

var (
	RedisClient redis.Client
)

func init() {
	//RedisClient = asynq.RedisClientOpt{
	//	Addr: "localhost:6379", // Redis server address
	//}
}
