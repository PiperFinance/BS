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
	"github.com/go-redis/redis/v8"

	"github.com/ethereum/go-ethereum"
	"github.com/hibiken/asynq"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BlockScanTaskHandler Uses BlockScanKey and requires no arg
// Start Scanning For new blocks -> enqueues a new fetch block task at the end
func BlockScanTaskHandler(ctx context.Context, task *asynq.Task) error {
	// blockTask := schema.BlockTask{ChainId: utils.BytesToInt(task.Payload())}
	blockTask := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		conf.Logger.Errorf("Task BlockScan [%+v] : %s ", blockTask, err)
		return err
	}
	err = blockScanTask(ctx, blockTask, *conf.QueueClient)
	_ = task
	if err != nil {
		conf.Logger.Errorf("Task BlockScan [%+v] : %s ", blockTask, err)
	} else {
		conf.Logger.Infof("Task BlockScan [%+v] ", blockTask)
	}
	return err
}

// BlockEventsTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func BlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", blockTask, err)
		return err
	}
	err = blockEventsTask(
		ctx,
		blockTask,
		*conf.QueueClient,
		*conf.GetMongoCol(blockTask.ChainId, conf.LogColName),
		blockTask.BlockNumber)
	if err != nil {
		conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", blockTask, err)
		return err
	}

	bm := schema.BlockM{BlockNumber: blockTask.BlockNumber}
	bm.SetFetched()
	if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
		ctx,
		bson.M{"no": blockTask.BlockNumber}, &bm); err != nil {
		conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", blockTask, err)
	} else {
		// conf.Logger.Infof("Replace Result : %d modified", res.ModifiedCount)
	}
	err = enqueuer.EnqueueParseBlockJob(*conf.QueueClient, blockTask)
	if err != nil {
		conf.Logger.Errorf("Task BlockEvents [%+v] : %s ", blockTask, err)
	} else {
		conf.Logger.Infof("Task BlockEvents [%+v] ", blockTask)
	}
	return err
}

// ParseBlockEventsTaskHandler Uses ParseBlockEventsKey and requires BlockTask as arg
// Parses Newly fetched events
func ParseBlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	blockTask := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
		return err
	}
	ctxFind, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cursor, err := conf.GetMongoCol(blockTask.ChainId, conf.LogColName).Find(ctxFind, bson.M{"blockNumber": &blockTask.BlockNumber})
	defer cursor.Close(ctxFind)
	if err != nil {
		return err
	}
	parsedLogsCol := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName)
	events.ParseLogs(ctx, parsedLogsCol, cursor)
	ctxDel, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = conf.GetMongoCol(blockTask.ChainId, conf.LogColName).DeleteMany(ctxDel, bson.M{"blockNumber": &blockTask.BlockNumber})
	if err != nil {
		return err
	}

	bm := schema.BlockM{BlockNumber: blockTask.BlockNumber, ChainId: blockTask.ChainId}
	bm.SetParsed()
	if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
		ctx,
		bson.M{"no": blockTask.BlockNumber}, &bm); err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
		return err
	} else {
		// conf.Logger.Infof("Replace Result : %d modified", res.ModifiedCount)
	}

	err = enqueuer.EnqueueUpdateUserBalJob(*conf.QueueClient, blockTask)
	if err != nil {
		conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", blockTask, err)
	} else {
		conf.Logger.Infof("Task ParseBlockEvents [%+v]", blockTask)
	}

	return err
}

// blockScanTask Enqueues a task to fetch events if new block is Found
func blockScanTask(ctx context.Context, blockTask schema.BlockTask, aqCl asynq.Client) error {
	ethCl := conf.EthClient(blockTask.ChainId)
	chain := blockTask.ChainId
	currentBlock, err := ethCl.BlockNumber(ctx)
	if conf.CallCount != nil {
		conf.CallCount.Add(blockTask.ChainId)
	}
	currentBlock -= conf.Config.BlockHeadDelay
	var lastBlock uint64

	if err != nil {
		conf.Logger.Errorf("Task BlockScan [%+v] : %s ", blockTask, err)
		return err
	}
	if lastBlockVal := conf.RedisClient.Get(ctx, tasks.LastScannedBlockKey(chain)); lastBlockVal.Err() == redis.Nil {
		lastBlock = conf.StartingBlock(ctx, blockTask.ChainId)
		if status := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), lastBlock, 0); status != nil && status.Err() != nil {
			conf.Logger.Errorf("Task BlockScan [%+v] : %s ", blockTask, err)
		}
	} else {
		if r, parseErr := lastBlockVal.Int(); parseErr != nil {
			conf.Logger.Errorf("blockScanTask: %s \nPossible issue is that somethings overwrote %s's value", parseErr, tasks.LastScannedBlockKey(chain))
			return err
		} else {
			lastBlock = uint64(r)
		}
	}
	if lastBlock < currentBlock {
		for blockNum := lastBlock; blockNum < currentBlock; blockNum++ {
			b := schema.BlockM{BlockNumber: blockNum, ChainId: chain}
			b.SetScanned()
			conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).InsertOne(ctx, &b)
			conf.NewBlockCount.Add(chain)
			_err := enqueuer.EnqueueFetchBlockJob(aqCl, schema.BlockTask{BlockNumber: b.BlockNumber, ChainId: chain})
			if _err != nil {
				return _err
			}
		}
		status := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), currentBlock, 0)
		if status != nil && status.Err() != nil {
			conf.Logger.Errorf("Task BlockScan [%+v] : %s ", blockTask, err)
		}
	}
	return err
}

// BlockEventsTask Fetches Block Events and stores them to mongo and enqueues another task for parsing them
func blockEventsTask(
	ctx context.Context,
	blockTask schema.BlockTask,
	aqCl asynq.Client,
	monCl mongo.Collection,
	blockNum uint64,
) error {
	blockNumBigInt := big.NewInt(int64(blockNum))
	logs, err := conf.EthClient(blockTask.ChainId).FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: blockNumBigInt,
			ToBlock:   blockNumBigInt,
		},
	)
	if err != nil {
		return err
	}
	if conf.CallCount != nil {
		conf.CallCount.Add(blockTask.ChainId)
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
	_, err = monCl.InsertMany(ctx, convLogs)
	if err != nil {
		return err
	}
	_ = aqCl
	return err
}
