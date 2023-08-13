package handlers

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/hibiken/asynq"
)

// FetchBlockTaskHandlerDummie
func FetchBlockTaskHandlerDummie(ctx context.Context, task *asynq.Task) error {
	bt := schema.BatchBlockTask{}
	if err := json.Unmarshal(task.Payload(), &bt); err != nil {
		return err
	}

	for i := bt.FromBlockNumber; i <= bt.ToBlockNumber; i++ {
		conf.Logger.Infow("Fetch ", "Block", i)
	}
	return nil
}

// FetchBlockTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func FetchBlockTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	if err := json.Unmarshal(task.Payload(), &blockTask); err != nil {
		return err
	}
	if err := fetchBlockEventsJob(ctx, blockTask); err != nil {
		return err
	}
	if err := enqueuer.EnqueueParseBlockJob(*conf.QueueClient, blockTask); err != nil {
		return err
	}
	return nil
}

// fetchBlockEventsJob Fetches Block Events and stores them to mongo and enqueues another task for parsing them
func fetchBlockEventsJob(
	ctx context.Context,
	blockTask schema.BatchBlockTask,
) error {
	// TODO: Retry With reduced range is this fails
	monCl := conf.GetMongoCol(blockTask.ChainId, conf.LogColName)
	logs, err := conf.EthClient(blockTask.ChainId).FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(blockTask.FromBlockNumber)),   // gte
			ToBlock:   big.NewInt(int64(blockTask.ToBlockNumber) - 1), // lt
		},
	)
	// NOTE: DEUBG
	conf.CallCount.Add(blockTask.ChainId)
	if err != nil {
		conf.FailedCallCount.Add(blockTask.ChainId)
		return &utils.RpcError{Err: err, ChainId: blockTask.ChainId, ToBlockNumber: blockTask.ToBlockNumber, FromBlockNumber: blockTask.FromBlockNumber, Name: "BlockFetch"}
	}

	if len(logs) < 1 {
		return nil
	}
	for i := blockTask.FromBlockNumber; i <= blockTask.ToBlockNumber; i++ {
		conf.Logger.Infow("Fetched", "block", i)
	}
	convLogs := make([]interface{}, 0)
	for _, _log := range logs {
		if !_log.Removed {
			convLogs = append(convLogs, schema.LogColl{
				Address:     _log.Address,
				Data:        _log.Data,
				Index:       _log.Index,
				Topics:      _log.Topics,
				TxIndex:     _log.TxIndex,
				BlockNumber: _log.BlockNumber,
				BlockHash:   _log.BlockHash,
				Removed:     _log.Removed,
				TxHash:      _log.TxHash,
			})
		}
	}
	if len(convLogs) > 0 {
		_, err = monCl.InsertMany(ctx, convLogs)
		if err != nil {
			return err
		}
	}

	return err
}
