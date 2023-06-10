package enqueuer

import (
	"encoding/json"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/hibiken/asynq"
)

const (
	BlockTaskGroup = "Block"
)

func EnqueueFetchBlockJob(aqCl asynq.Client, blockTask schema.BatchBlockTask) error {
	payload, err := json.Marshal(blockTask)
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.FetchBlockEventsKey, payload),
		asynq.Queue(conf.FetchQ),
		asynq.Timeout(conf.Config.FetchBlockTimeout),
	)
	return err
}

func EnqueueParseBlockJob(aqCl asynq.Client, blockTask schema.BatchBlockTask) error {
	payload, err := json.Marshal(blockTask)
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.ParseBlockEventsKey, payload),
		asynq.Queue(conf.ParseQ),
		asynq.Timeout(conf.Config.ParseBlockTimeout),
	)
	return err
}
