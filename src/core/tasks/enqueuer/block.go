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

func EnqueueProcessBlockJob(aqCl asynq.Client, blockTask schema.BatchBlockTask) error {
	payload, err := json.Marshal(blockTask)
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.ProccessBlockKey, payload),
		asynq.Queue(conf.ProcessQ),
		asynq.Timeout(conf.Config.ProcessBlockTimeout),
	)
	return err
}

// EnqueueFetchBlockJob This function is deprecated in favour of a faster all in one flow
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

// EnqueueParseBlockJob  This function is deprecated in favour of a faster all in one flow
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
