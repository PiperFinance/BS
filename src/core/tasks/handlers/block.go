package handlers

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/go-redis/redis/v8"

	"github.com/ethereum/go-ethereum"
	"github.com/hibiken/asynq"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BlockScanTaskHandler Uses BlockScanKey and requires no arg
// Start Scanning For new blocks -> enqueues a new fetch block task at the end
func BlockScanTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		return err
	}
	err = blockScanTask(ctx, blockTask, *conf.QueueClient)
	_ = task
	return err
}

// BlockEventsTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func BlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		// conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", blockTask, err)
		return err
	}
	err = blockEventsTask(
		ctx,
		blockTask,
		*conf.QueueClient,
		*conf.GetMongoCol(blockTask.ChainId, conf.LogColName),
	)
	if err != nil {
		// conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", blockTask, err)
		return err
	}
	for i := blockTask.FromBlockNumber; i < blockTask.ToBlockNumber; i++ {
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
	ctxFind, cancelFind := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelFind()
	ctxDel, cancelDel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelDel()
	filter := bson.M{"blockNumber": bson.D{{Key: "$gte", Value: blockTask.FromBlockNumber}, {Key: "$lt", Value: blockTask.ToBlockNumber}}}
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
	_, err = conf.GetMongoCol(blockTask.ChainId, conf.LogColName).DeleteMany(ctxDel, filter)
	if err != nil {
		return err
	}
	for i := blockTask.FromBlockNumber; i < blockTask.ToBlockNumber; i++ {

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

// blockScanTask Enqueues a task to fetch events if new block is Found
func blockScanTask(ctx context.Context, blockTask schema.BatchBlockTask, aqCl asynq.Client) error {
	var lastBlock uint64
	chain := blockTask.ChainId
	currentBlock, err := conf.LatestBlock(ctx, blockTask.ChainId)
	if err != nil {
		return &utils.RpcError{Err: err, ChainId: blockTask.ChainId, ToBlockNumber: blockTask.ToBlockNumber, FromBlockNumber: blockTask.FromBlockNumber, Name: "BlockScan"}
	}

	if lastBlockVal := conf.RedisClient.Get(ctx, tasks.LastScannedBlockKey(chain)); lastBlockVal.Err() == redis.Nil {
		// NOTE - First Running no head block is stored
		// if _lastBlock, err := conf.LatestBlock(ctx, blockTask.ChainId); err != nil {
		// 	return &utils.RpcError{Err: err, ChainId: blockTask.ChainId, ToBlockNumber: blockTask.ToBlockNumber, FromBlockNumber: blockTask.FromBlockNumber, Name: "BlockScan"}
		// } else {
		lastBlock = currentBlock
		if status := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), lastBlock, 0); status != nil && status.Err() != nil {
			return status.Err()
		}
		// lastBlock = _lastBlock
		// }
	} else {
		// NOTE - Next Iterations there is head block and is parsed below
		if r, parseErr := lastBlockVal.Int(); parseErr != nil {
			conf.Logger.Errorf("blockScanTask: %s \nPossible issue is that somethings overwrote %s's value", parseErr, tasks.LastScannedBlockKey(chain))
			return err
		} else {
			lastBlock = uint64(r)
		}
	}
	batchSize := conf.BatchLogMaxHeight(chain)
	head := lastBlock + batchSize
	// TODO Test here !
	if head < currentBlock {
		remainingBlocks := currentBlock - lastBlock
		batchCount := remainingBlocks / batchSize
		newBlocks := make([]interface{}, remainingBlocks)
		i := 0
		for blockNum := lastBlock; blockNum < currentBlock; blockNum++ {
			b := schema.BlockM{BlockNumber: blockNum, ChainId: chain}
			b.SetScanned()
			newBlocks[i] = b
			i++
		}
		var j uint64
		for j = 0; j <= batchCount; j++ {
			b := schema.BatchBlockTask{
				FromBlockNumber: lastBlock + (j * batchSize),
				ToBlockNumber:   lastBlock + ((j + 1) * batchSize),
				ChainId:         chain,
			}
			if b.ToBlockNumber > currentBlock {
				b.ToBlockNumber = currentBlock
			}
			_err := enqueuer.EnqueueFetchBlockJob(aqCl, b)
			if _err != nil {
				return _err
			}
		}
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).InsertMany(ctx, newBlocks); err != nil {
			return err
		}
		status := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), currentBlock, 0)
		if status != nil && status.Err() != nil {
			return status.Err()
		}
		conf.NewBlockCount.AddFor(chain, uint64(len(newBlocks)))
	}
	return err
}

// BlockEventsTask Fetches Block Events and stores them to mongo and enqueues another task for parsing them
func blockEventsTask(
	ctx context.Context,
	blockTask schema.BatchBlockTask,
	aqCl asynq.Client,
	monCl mongo.Collection,
) error {
	// TODO - Retry With reduced range is this fails
	logs, err := conf.EthClient(blockTask.ChainId).FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(blockTask.FromBlockNumber)),
			ToBlock:   big.NewInt(int64(blockTask.ToBlockNumber)),
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
	convLogs := make([]interface{}, len(logs))
	for i, _log := range logs {
		convLogs[i] = schema.LogColl{
			Address:     _log.Address,
			Data:        _log.Data,
			Index:       _log.Index,
			Topics:      _log.Topics,
			TxIndex:     _log.TxIndex,
			BlockNumber: _log.BlockNumber,
			BlockHash:   _log.BlockHash,
			Removed:     _log.Removed,
			TxHash:      _log.TxHash,
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
