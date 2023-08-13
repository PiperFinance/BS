package handlers

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
	"github.com/PiperFinance/BS/src/utils"
)

// BlockScanTaskHandler Uses BlockScanKey and requires no arg
// Start Scanning For new blocks -> enqueues a new fetch block task at the end
func BlockScanTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		return err
	}
	err = scanBlockJob(ctx, blockTask, *conf.QueueClient)
	_ = task
	return err
}

// BlockEventsTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func BlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		return err
	}
	err = fetchBlockEventsJob(
		ctx,
		blockTask,
		*conf.QueueClient,
		*conf.GetMongoCol(blockTask.ChainId, conf.LogColName),
	)
	if err != nil {
		return err
	}
	for i := blockTask.FromBlockNumber; i <= blockTask.ToBlockNumber; i++ {
		bm := schema.BlockM{BlockNumber: i}
		bm.SetFetched()
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
			ctx,
			bson.M{"no": i}, &bm); err != nil {
			// Does not Stop here Since this is not a very important err
			conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", bm, err)
		}
	}
	if err := enqueuer.EnqueueParseBlockJob(*conf.QueueClient, blockTask); err != nil {
		return err
	}
	return err
}

// ParseBlockEventsTaskHandler Uses ParseBlockEventsKey and requires BlockTask as arg
// Parses Newly fetched events
func ParseBlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
		return err
	}
	ctxFind, cancelFind := context.WithTimeout(ctx, conf.Config.MongoMaxTimeout)
	defer cancelFind()

	filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: blockTask.FromBlockNumber}, {Key: "$lte", Value: blockTask.ToBlockNumber}}}

	cursor, err := conf.GetMongoCol(blockTask.ChainId, conf.LogColName).Find(ctxFind, filter)
	defer func() {
		if err := cursor.Close(ctxFind); err != nil {
			conf.Logger.Error(err)
		}
	}()
	if err != nil {
		return err
	}
	parsedLogsCol := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName)
	events.ParseLogs(ctx, parsedLogsCol, cursor)
	for i := blockTask.FromBlockNumber; i < blockTask.ToBlockNumber; i++ {
		conf.Logger.Infow("Parsed", "block", i)
	}

	// This is removed in favour of another job vaccumming logs (task advantage of logCol index in mongo)
	// ctxDel, cancelDel := context.WithTimeout(ctx, conf.Config.MongoMaxTimeout)
	// filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: blockTask.FromBlockNumber}, {Key: "$lt", Value: blockTask.ToBlockNumber}}}
	// defer cancelDel()
	// _, err = conf.GetMongoCol(blockTask.ChainId, conf.LogColName).DeleteMany(ctxDel, filter)
	err = conf.RedisClient.SetLogsToVaccum(ctx, blockTask.ChainId, blockTask.FromBlockNumber, blockTask.ToBlockNumber)
	if err != nil {
		return err
	}

	for i := blockTask.FromBlockNumber; i <= blockTask.ToBlockNumber; i++ {
		bm := schema.BlockM{BlockNumber: i, ChainId: blockTask.ChainId}
		bm.SetParsed()
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
			ctx,
			bson.M{"no": i}, &bm); err != nil {
			conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
			return err
		}
	}
	// TODO - Enqueue Other Tasks !
	err = enqueuer.EnqueueUpdateUserBalJob(*conf.QueueClient, blockTask)
	if err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
	} else {
		conf.Logger.Infof("Task ParseBlockEvents [%+v]", blockTask)
	}

	return err
}

// scanBlockJob Enqueues a task to fetch events if new block range is met the  height
// from head up current block - delay
// 🚧 inner app queries are gte -> lt
func scanBlockJob(ctx context.Context, blockTask schema.BatchBlockTask, aqCl asynq.Client) error {
	var head uint64
	chain := blockTask.ChainId
	batchSize := conf.BatchLogMaxHeight(chain)
	currentBlock, err := conf.LatestBlock(ctx, blockTask.ChainId)
	if err != nil {
		return &utils.RpcError{Err: err, ChainId: blockTask.ChainId, ToBlockNumber: blockTask.ToBlockNumber, FromBlockNumber: blockTask.FromBlockNumber, Name: "BlockScan"}
	}

	if cmd := conf.RedisClient.Get(ctx, tasks.LastScannedBlockKey(chain)); cmd.Err() != nil {
		return cmd.Err()
	} else {
		// NOTE - Next Iterations there is head block and is parsed below
		if r, parseErr := cmd.Int(); parseErr != nil {
			conf.Logger.Errorf("blockScanTask: %s \nPossible issue is that somethings overwrote %s's value", parseErr, tasks.LastScannedBlockKey(chain))
			return err
		} else {
			head = uint64(r)
		}
	}
	// FIXME: remove please
	conf.Logger.Infow("BlockScan", "block", currentBlock, "head", head, "b-size", batchSize)

	if head+batchSize < currentBlock {
		newHead := head
		newBlocks := make([]interface{}, 0)
		for newHead < currentBlock {
			b := schema.BlockM{BlockNumber: newHead, ChainId: chain}
			b.SetScanned()
			newBlocks = append(newBlocks, b)
			newHead++
		}

		batchCount := (newHead - head) / batchSize
		if batchCount == 0 {
			return nil
		}

		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).InsertMany(ctx, newBlocks); err != nil {
			return err
		}

		for j := uint64(0); j < batchCount; j += 2 {
			b := schema.BatchBlockTask{
				FromBlockNumber: head + (j * batchSize),
				ToBlockNumber:   head + ((j + 1) * batchSize),
				ChainId:         chain,
			}
			if b.ToBlockNumber > currentBlock {
				b.ToBlockNumber = currentBlock
			}
			if b.FromBlockNumber == b.ToBlockNumber {
				break
			}
			if _err := enqueuer.EnqueueFetchBlockJob(aqCl, b); _err != nil {
				return _err
			}
			for i := b.FromBlockNumber; i <= b.ToBlockNumber; i++ {
				conf.Logger.Infow("Enqueue Scan", "block", i)
			}

			conf.NewBlockCount.AddFor(chain, uint64(b.ToBlockNumber-b.FromBlockNumber))
		}

		if cmd := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), newHead, 0); cmd.Err() != nil {
			return cmd.Err()
		}
	}
	return err
}

// fetchBlockEventsJob Fetches Block Events and stores them to mongo and enqueues another task for parsing them
func fetchBlockEventsJob(
	ctx context.Context,
	blockTask schema.BatchBlockTask,
	aqCl asynq.Client,
	monCl mongo.Collection,
) error {
	// TODO - Retry With reduced range is this fails
	logs, err := conf.EthClient(blockTask.ChainId).FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(blockTask.FromBlockNumber)),   // gte
			ToBlock:   big.NewInt(int64(blockTask.ToBlockNumber) - 1), // lt
		},
	)
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
	_ = aqCl
	return err
}
