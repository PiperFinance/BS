package handlers

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/jobs"
	"github.com/hibiken/asynq"
)

// ProccessBlockTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func ProccessBlockTaskHandler(ctx context.Context, task *asynq.Task) error {
	// TODO: Add check for task retries X_X

	defer func() {
		if r := recover(); r != nil {
			conf.Logger.Infof("stack trace from panic: %s", string(debug.Stack()))
		}
	}()

	bt := schema.BatchBlockTask{}
	if err := json.Unmarshal(task.Payload(), &bt); err != nil {
		return err
	}
	err, logs := jobs.FetchBlockEvents(ctx, bt)
	if err != nil {
		return err
	}
	helpers.SetBTFetched(ctx, bt)

	err = jobs.ProccessLogs(ctx, bt, logs)
	if err != nil {
		return err
	}
	helpers.SetBTParsed(ctx, bt)

	// err = jobs.PrcoessParsedLogs(ctx, bt, transfers)
	// if err != nil {
	// 	return err
	// }
	helpers.SetBTAdded(ctx, bt)

	return nil
}
