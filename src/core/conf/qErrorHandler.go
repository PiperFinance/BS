package conf

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

type QueueErrorHandler struct{}

func (er *QueueErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	retried, _ := asynq.GetRetryCount(ctx)
	maxRetry, _ := asynq.GetMaxRetry(ctx)
	if retried >= maxRetry {
		// err = fmt.`Errorf("retry exhausted for task %s: %w", task.Type, err)
		insertCtx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()
		MongoDB.Collection(QueueErrorsColName).InsertOne(insertCtx, bson.M{
			"time": time.Now().Unix(),
			"task": bson.M{
				"payload": task.Payload(),
				"type":    task.Type(),
			},
			"err":      err.Error(),
			"retries":  retried,
			"maxRetry": maxRetry,
		})
	}
	// errorReportingService.Notify(err)
}
