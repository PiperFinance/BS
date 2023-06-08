package conf

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/hibiken/asynq"
)

type QueueErrorHandler struct{}

func errType(ChainId int64, v interface{}) interface{} {
	switch v.(type) {
	case error:
		return v
	case utils.RpcError:
		FailedCallCount.Add(ChainId)
		if Config.SilenceRRCErrs {
			return nil
		} else {
			return v
		}
	default:
		return "unknown"
	}
}

func (er *QueueErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	retried, _ := asynq.GetRetryCount(ctx)
	blockTask := schema.BlockTask{}
	if errJson := json.Unmarshal(task.Payload(), &blockTask); errJson == nil && blockTask.ChainId > 0 {
		if errType(blockTask.ChainId, err) == nil {
			return
		}
		Logger.Errorw("QErr", "Retires", retried, "block", blockTask, "err", err)
	} else {
		Logger.Errorw("QErr", "task", task.Type(), "Retries", retried, "err", err)
	}
}
