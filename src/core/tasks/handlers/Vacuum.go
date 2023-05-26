package handlers

import (
	"context"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/hibiken/asynq"
)

func VacuumLogHandler(ctx context.Context, task *asynq.Task) error {
	// Get LastVacuumed Block number
	conf.MongoDB.Collection(conf.LogColName)
	// start int(getLastBlock() - conf.VacuumLogsHeight)
	// for i := ; i < ; i++ {

	// }getLastBlock()

	// // Save LastVacuumed BlockNumber
	_, _ = ctx, task
	return nil
}
