package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/tasks"
)

// if res := conf.RedisClient.Get(c.Context(), tasks.LastScannedBlockKey); res.Err() == nil {
// 	lb, err := res.Uint64()
// 	if err != nil {
// 		return err
// 	} else {
// 		lastBlock = lb
// 	}
// }

func GetLastBlock(chain int64) (uint64, error) {
	var lastBlock uint64
	ctx := context.TODO()
	c, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	if res := conf.RedisClient.Get(c, tasks.LastScannedBlockKey(chain)); res.Err() != nil {
		return lastBlock, res.Err()
	} else {
		val, castErr := res.Uint64()
		lastBlock = val
		if castErr != nil {
			return lastBlock, fmt.Errorf("Can not cast to %d uint , err: %s", lastBlock, castErr)
		}
	}
	return lastBlock, nil
}
