package conf

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/hibiken/asynq"
)

type QueueErrorHandler struct{}

func (er *QueueErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	if err == nil {
		return
	}
	retried, _ := asynq.GetRetryCount(ctx)
	// maxRetry, _ := asynq.GetMaxRetry(ctx)
	// var errCol *mongo.Collection
	blockTask := schema.BlockTask{}
	if errJson := json.Unmarshal(task.Payload(), &blockTask); errJson == nil && blockTask.ChainId > 0 {
		// errCol = GetMongoCol(blockTask.ChainId, QueueErrorsColName)
		Logger.Errorf("Retries:%d [%d] @ %d : %+v", retried, blockTask.ChainId, blockTask.BlockNumber, err)
	} else {
		// errCol = MongoDefaultErrCol
		Logger.Errorf("Retries:%d : %+v", retried, err)
	}

	// insertCtx, cancel := context.WithTimeout(context.TODO(), time.Second)
	// defer cancel()

	// errCol.InsertOne(insertCtx, bson.M{
	// 	"time": time.Now().Unix(),
	// 	"task": bson.M{
	// 		"payload":   task.Payload(),
	// 		"blockTask": blockTask,
	// 		"type":      task.Type(),
	// 	},
	// 	"err":      err.Error(),
	// 	"retries":  retried,
	// 	"maxRetry": maxRetry,
	// })

	// if retried >= maxRetry {
	// }
	// errorReportingService.Notify(err)
}
