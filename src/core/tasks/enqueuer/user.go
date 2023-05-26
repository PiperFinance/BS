package enqueuer

import (
	"encoding/json"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/hibiken/asynq"
)

func EnqueueUpdateUserBalJob(aqCl asynq.Client, blockTask schema.BlockTask) error {
	payload, err := json.Marshal(blockTask)
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.UpdateUserBalanceKey, payload),
		asynq.Queue(conf.ProcessQ),
		asynq.Timeout(conf.Config.UserBalUpdateTimeout))
	return err
}
