package enqueuer

import (
	"encoding/json"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/hibiken/asynq"
)

func EnqueueFetchBlockJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(schema.BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(asynq.NewTask(tasks.FetchBlockEventsKey, payload), asynq.Queue(conf.FetchQ))
	return err
}

func EnqueueParseBlockJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(schema.BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(asynq.NewTask(tasks.ParseBlockEventsKey, payload), asynq.Queue(conf.ParseQ))
	return err
}

func EnqueueUpdateUserBalJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(schema.BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(asynq.NewTask(tasks.UpdateUserBalanceKey, payload), asynq.Queue(conf.ProcessQ))
	return err
}
