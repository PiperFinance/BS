package conf

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/hibiken/asynq"
)

type QueueErrorHandler struct{}

func errType(v interface{}) interface{} {
	switch v.(type) {
	case error:
		return v
	case utils.RpcError:
		if Config.SilenceRRCErrs {
			return v
		} else {
			return nil
		}
	default:
		return "unknown"
	}
}

func (er *QueueErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	if errType(err) == nil {
		return
	}
	retried, _ := asynq.GetRetryCount(ctx)
	blockTask := schema.BlockTask{}
	if errJson := json.Unmarshal(task.Payload(), &blockTask); errJson == nil && blockTask.ChainId > 0 {
		Logger.Errorf("Retries:%d [%d] @ %d : %+v", retried, blockTask.ChainId, blockTask.BlockNumber, err)
	} else {
		Logger.Errorf("Retries:%d : %+v", retried, err)
	}
}
