package handlers

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hibiken/asynq"
)

func ProccessBlockTaskHandler(ctx context.Context, task *asynq.Task) error {
	bt := schema.BatchBlockTask{}
	if err := json.Unmarshal(task.Payload(), &bt); err != nil {
		return err
	}
	err, logs := fetchBlockEventsJob(ctx, bt)
	if err != nil {
		return err
	}
	err, transfers := proccessRawLogs(ctx, bt, logs)
	if err != nil {
		return err
	}

	err = updateUserBalJob(ctx, bt, transfers)
	if err != nil {
		return err
	}
	helpers.SetBTFetched(ctx, bt)
	if err := enqueuer.EnqueueParseBlockJob(*conf.QueueClient, bt); err != nil {
		return err
	}
	checkResults(ctx, bt)
	return nil
}

// proccessRawLogs [TEMP] Parses raw logs and of they are transfer logs stores them ??
func proccessRawLogs(ctx context.Context, bt schema.BatchBlockTask, blocksLogs map[uint64][]types.Log) (error, map[uint64][]schema.LogTransfer) {
	res := make(map[uint64][]schema.LogTransfer)
	for blockNo, logs := range blocksLogs {
		blockTrxs := make([]schema.LogTransfer, 0)
		for _, log := range logs {
			parsedLog, err := events.ParseLog(log)
			if err != nil {
				switch err.(type) {
				case *utils.ErrEventParserNotFound:
					if !conf.Config.SilenceParseErrs {
						conf.Logger.Errorw("ParseLogs", "err", err)
					}
					continue
					// case error:
					// 	return err, nil
				}
			}
			if parsedLog == nil {
				continue
			}

			// TODO: this should be a switch case type
			trxLog, ok := parsedLog.(schema.LogTransfer)
			if !ok {
				continue
			}
			blockTrxs = append(blockTrxs, trxLog)
		}
		b := schema.BlockTask{ChainId: bt.ChainId, BlockNumber: blockNo}
		if err := processTransferLogs(ctx, b, blockTrxs); err != nil {
			return err, nil
		}
		res[blockNo] = blockTrxs
	}
	return nil, res
}
