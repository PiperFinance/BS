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
	bt := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &bt)
	if err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", bt, err)
		return err
	}
	parseNewBlocks(ctx, bt)
	return nil
}

func parseNewBlocks(ctx context.Context, bt schema.BatchBlockTask) error {
	ctxFind, cancelFind := context.WithTimeout(ctx, conf.Config.MongoMaxTimeout)
	defer cancelFind()

	filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: bt.FromBlockNum}, {Key: "$lte", Value: bt.ToBlockNum}}}

	cursor, err := conf.GetMongoCol(bt.ChainId, conf.LogColName).Find(ctxFind, filter)
	defer func() {
		if err := cursor.Close(ctxFind); err != nil {
			conf.Logger.Error(err)
		}
	}()
	if err != nil {
		return err
	}
	events.ParseLogs(ctx, bt, cursor)
	for i := bt.FromBlockNum; i < bt.ToBlockNum; i++ {
		conf.Logger.Infow("Parsed", "block", i)
	}

	if err := conf.RedisClient.SetRawLogsToVaccum(ctx, bt.ChainId, bt.FromBlockNum, bt.ToBlockNum); err != nil {
		return err
	}

	// TODO:  Enqueue Other Tasks !
	if err := enqueuer.EnqueueUpdateUserBalJob(*conf.QueueClient, bt); err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", bt, err)
	} else {
		conf.Logger.Infof("Task ParseBlockEvents [%+v]", bt)
	}

	helpers.SetBTParsed(ctx, bt)
	return err
}
