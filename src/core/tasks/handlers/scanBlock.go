package handlers

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/jobs"
	"github.com/hibiken/asynq"
)

// BlockScanTaskHandler Uses Block Scan Key and requires no arg
// Start Scanning For new blocks -> enqueues a new fetch block task at the end
func BlockScanTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		return err
	}
	err = jobs.ScanBlockJob(ctx, blockTask, *conf.QueueClient)
	_ = task
	return err
}
