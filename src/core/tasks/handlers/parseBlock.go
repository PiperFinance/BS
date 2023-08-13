package handlers

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
)

// ParseBlockEventsTaskHandler Uses ParseBlockEventsKey and requires BlockTask as arg
// Parses Newly fetched events
func ParseBlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
		return err
	}
	ctxFind, cancelFind := context.WithTimeout(ctx, conf.Config.MongoMaxTimeout)
	defer cancelFind()

	filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: blockTask.FromBlockNumber}, {Key: "$lte", Value: blockTask.ToBlockNumber}}}

	cursor, err := conf.GetMongoCol(blockTask.ChainId, conf.LogColName).Find(ctxFind, filter)
	defer func() {
		if err := cursor.Close(ctxFind); err != nil {
			conf.Logger.Error(err)
		}
	}()
	if err != nil {
		return err
	}
	parsedLogsCol := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName)
	events.ParseLogs(ctx, parsedLogsCol, cursor)
	for i := blockTask.FromBlockNumber; i < blockTask.ToBlockNumber; i++ {
		conf.Logger.Infow("Parsed", "block", i)
	}

	if err := conf.RedisClient.SetRawLogsToVaccum(ctx, blockTask.ChainId, blockTask.FromBlockNumber, blockTask.ToBlockNumber); err != nil {
		return err
	}

	// TODO:  Enqueue Other Tasks !
	if err := enqueuer.EnqueueUpdateUserBalJob(*conf.QueueClient, blockTask); err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
	} else {
		conf.Logger.Infof("Task ParseBlockEvents [%+v]", blockTask)
	}

	helpers.SetBTParsed(ctx, blockTask)
	return err
}
