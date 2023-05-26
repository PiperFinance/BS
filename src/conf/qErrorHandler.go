package conf

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type QueueErrorHandler struct{}

func (er *QueueErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	var errCol *mongo.Collection
	blockTask := schema.BlockTask{}
	if err := json.Unmarshal(task.Payload(), &blockTask); err == nil && blockTask.ChainId > 0 {
		errCol = GetMongoCol(blockTask.ChainId, QueueErrorsColName)
	} else {
		errCol = MongoDefaultErrCol
	}

	insertCtx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	errCol.InsertOne(insertCtx, bson.M{
		"time": time.Now().Unix(),
		"task": bson.M{
			"payload":   task.Payload(),
			"blockTask": blockTask,
			"type":      task.Type(),
		},
		"err":      err.Error(),
		"retries":  retried,
		"maxRetry": maxRetry,
	})
	// if retried >= maxRetry {
	// }
	// errorReportingService.Notify(err)
}
