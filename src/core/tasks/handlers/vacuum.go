package handlers

import (
	"context"

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
		vacRng, err := conf.RedisClient.GetRawLogsToVaccum(ctx, chain)
		if err != nil {
			return err
		} else if vacRng == nil {
			return nil
		} else {
			filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: vacRng.FromBlock}, {Key: "$lte", Value: vacRng.ToBlock}}}
			_, err := conf.GetMongoCol(chain, conf.LogColName).DeleteMany(ctx, filter)
			if err != nil {
				return err
			}
		}
	}
}

// vaccumParsedLogOID ObjectId Parsed Log that are read and saved into db (as approve, transfer , ... )
func vaccumParsedLogOID(ctx context.Context, chain int64) error {
	ids, err := conf.RedisClient.GetParsedLogsIDsToVaccum(ctx, chain)
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
	for _, chain := range conf.Config.SupportedChains {
		// NOTE: This should not be triggered since no log is being stored
		err := vaccumRawLogs(ctx, chain)
		if err != nil {
			conf.Logger.Errorw("vaccumRawLogs", "err", err, "chain", chain)
		}
		err = vaccumParsedLogOID(ctx, chain)
		if err != nil {
			conf.Logger.Errorw("vaccumParsedLogOID", "err", err, "chain", chain)
		}
	}
	return nil
}
