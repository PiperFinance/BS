package handlers

import (
	"context"
	"sync"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

func vaccumBlockRange(ctx context.Context, chain int64) error {
	for {
		vacRng, err := conf.RedisClient.GetLogsToVaccum(ctx, chain)
		if err != nil {
			return err
		} else if vacRng == nil {
			return nil
		} else {
			filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: vacRng.FromBlock}, {Key: "$lt", Value: vacRng.ToBlock}}}
			_, err := conf.GetMongoCol(chain, conf.ParsedLogColName).DeleteMany(ctx, filter)
			if err != nil {
				return err
			}
		}
	}
}

func vaccumObjIds(ctx context.Context, chain int64) error {
	ids, err := conf.RedisClient.GetLogsIDsToVaccum(ctx, chain)
	if err != nil {
		return err
	} else if ids == nil {
		return nil
	} else {
		_, err := conf.GetMongoCol(chain, conf.ParsedLogColName).DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
		if err != nil {
			return err
		}
		return nil
	}
}

func VacuumLogHandler(ctx context.Context, task *asynq.Task) error {
	// return nil
	wg := sync.WaitGroup{}
	wg.Add(2)
	for _, chain := range conf.Config.SupportedChains {
		go func(chain int64) {
			defer wg.Done()
			err := vaccumBlockRange(ctx, chain)
			if err != nil {
				conf.Logger.Errorw("vaccumBlockRange", "err", err, "chain", chain)
			}
		}(chain)
		go func(chain int64) {
			defer wg.Done()
			err := vaccumObjIds(ctx, chain)
			if err != nil {
				conf.Logger.Errorw("vaccumBlockRange", "err", err, "chain", chain)
			}
		}(chain)
	}
	wg.Wait()
	return nil
}
