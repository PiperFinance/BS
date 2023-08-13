package handlers

import (
	"context"
	"sync"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

// vaccumRawLogs Vaccum Logs that are parsed and are not longer needed
func vaccumRawLogs(ctx context.Context, chain int64) error {
	/*
		added in favour of
		// ctxDel, cancelDel := context.WithTimeout(ctx, conf.Config.MongoMaxTimeout)
		// filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: blockTask.FromBlockNumber}, {Key: "$lt", Value: blockTask.ToBlockNumber}}}
		// defer cancelDel()
		// _, err = conf.GetMongoCol(blockTask.ChainId, conf.LogColName).DeleteMany(ctxDel, filter)
	*/
	for {
		vacRng, err := conf.RedisClient.GetLogsToVaccum(ctx, chain)
		if err != nil {
			return err
		} else if vacRng == nil {
			return nil
		} else {
			filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: vacRng.FromBlock}, {Key: "$lt", Value: vacRng.ToBlock}}}
			_, err := conf.GetMongoCol(chain, conf.LogColName).DeleteMany(ctx, filter)
			if err != nil {
				return err
			}
		}
	}
}

// vaccumParsedLogOID Parsed Log that are read and saved into db (as approve, transfer , ... )
func vaccumParsedLogOID(ctx context.Context, chain int64) error {
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

// VacuumLogHandler Acts as entry point for periodic house cleaning periodic tasks
func VacuumLogHandler(ctx context.Context, task *asynq.Task) error {
	// return nil
	wg := sync.WaitGroup{}
	wg.Add(2)
	for _, chain := range conf.Config.SupportedChains {
		go func(chain int64) {
			defer wg.Done()
			err := vaccumRawLogs(ctx, chain)
			if err != nil {
				conf.Logger.Errorw("vaccumRawLogs", "err", err, "chain", chain)
			}
		}(chain)
		go func(chain int64) {
			defer wg.Done()
			err := vaccumParsedLogOID(ctx, chain)
			if err != nil {
				conf.Logger.Errorw("vaccumParsedLogOID", "err", err, "chain", chain)
			}
		}(chain)
	}
	wg.Wait()
	return nil
}
