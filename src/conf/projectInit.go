package conf

import (
	"context"
	"log"

	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/go-redis/redis/v8"
)

func LoadProjectInit() {
	ctx := context.TODO()
	for _, chain := range Config.SupportedChains {
		// TODO: Check if Mongo is empty or not
		k := tasks.LastScannedBlockKey(chain)
		if cmd := RedisClient.Get(ctx, k); cmd.Err() == redis.Nil {
			headBlock, err := LatestBlock(ctx, chain)
			if err != nil {
				log.Panic(err)
			}
			if cmd := RedisClient.Set(ctx, k, headBlock, 0); cmd.Err() != nil {
				log.Panic(err)
			}
		}
	}
}
