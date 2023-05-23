package enqueuer

import (
	"encoding/json"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/hibiken/asynq"
)

const (
	UserBalUpdateTimeout = 50 * time.Second
)

func EnqueueUpdateUserBalJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(schema.BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.UpdateUserBalanceKey, payload),
		asynq.Queue(conf.ProcessQ),
		asynq.Timeout(UserBalUpdateTimeout))
	return err
}
