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
	ParseBlockTimeout = 100 * time.Second
	FetchBlockTimeout = 2 * time.Minute
	BlockTaskGroup    = "Block"
)

func EnqueueFetchBlockJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(schema.BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.FetchBlockEventsKey, payload),
		asynq.Queue(conf.FetchQ),
		asynq.Timeout(FetchBlockTimeout),
		asynq.Group(BlockTaskGroup))
	return err
}

func EnqueueParseBlockJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(schema.BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(
		asynq.NewTask(tasks.ParseBlockEventsKey, payload),
		asynq.Queue(conf.ParseQ),
		asynq.Timeout(ParseBlockTimeout),
		asynq.Group(BlockTaskGroup))
	return err
}
