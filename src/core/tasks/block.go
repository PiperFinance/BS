package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// LastScannedBlockKey TODO - Add a model in db to save fetched block numbers
	// TODO - use gocache - not redis ...
	// MultiChain ...
	LastScannedBlockKey = "block:lastScanned"

	FetchBlockEventsKey  = "block:fetch_events"
	BlockScanKey         = "block:scan"
	ParseBlockEventsKey  = "block:parse_events"
	UpdateUserBalanceKey = "user:update_bal"

	VacuumLogsKey     = "chore:vacuum"
	VacuumLogsLockKey = "chore:vacuum-lock"
	VacuumLogsHeight  = 100

	// TypeBlockSearchKey  = "block:search"
)

type BlockTask struct {
	BlockNumber uint64
}

// BlockScanTaskHandler Uses BlockScanKey and requires no arg
func BlockScanTaskHandler(ctx context.Context, task *asynq.Task) error {
	err := blockScanTask(ctx, *conf.EthClient, *conf.QueueClient)
	_ = task
	if err != nil {
		log.Errorf("Task BlockScan [%s] : Finished !", err)
	} else {
		log.Infof("Task BlockScan [OK] : Finished !")
	}
	return err
}

// BlockEventsTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func BlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	block := BlockTask{}
	mongoCollection := conf.MongoDB.Collection(conf.LogColName)
	err := json.Unmarshal(task.Payload(), &block)
	if err != nil {
		log.Errorf("Task BlockEvents [%s] : Finished !", err)
		return err
	}
	err = blockEventsTask(ctx, *conf.EthClient, *conf.QueueClient, *mongoCollection, block.BlockNumber)
	if err != nil {
		log.Errorf("Task BlockEvents [%s] : Finished !", err)
		return err
	}
	err = enqueueParseBlockJob(*conf.QueueClient, block.BlockNumber)
	if err != nil {
		log.Errorf("Task BlockEvents [%d] : Err : %s !", block.BlockNumber, err)
	} else {
		log.Infof("Task BlockEvents [%d] : Finished !", block.BlockNumber)
	}
	return err
}

// ParseBlockEventsTaskHandler Uses ParseBlockEventsKey and requires BlockTask as arg
// Parses Newly fetched events
func ParseBlockEventsTaskHandler(ctx context.Context, task *asynq.Task) error {
	block := BlockTask{}
	mongoParsedLogsCol := conf.MongoDB.Collection(conf.ParsedLogColName)
	err := json.Unmarshal(task.Payload(), &block)
	if err != nil {
		log.Infof("Task ParseBlockEvents [%s] : Finished !", err)
		return err
	}
	ctxFind, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cursor, err := conf.MongoDB.Collection(conf.LogColName).Find(ctxFind, bson.M{"blockNumber": &block.BlockNumber})
	defer cursor.Close(ctxFind)
	if err != nil {
		return err
	}
	events.ParseLogs(ctx, mongoParsedLogsCol, cursor)
	ctxDel, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = conf.MongoDB.Collection(conf.LogColName).DeleteMany(ctxDel, bson.M{"blockNumber": &block.BlockNumber})
	if err != nil {
		return err
	}
	err = enqueueUpdateUserBalJob(*conf.QueueClient, block.BlockNumber)

	if err != nil {
		log.Errorf("Task ParseBlockEvents [%d] : Err : %s !", block.BlockNumber, err)
	} else {
		log.Infof("Task ParseBlockEvents [%d] : Finished !", block.BlockNumber)
	}

	return err
}

// blockScanTask Enqueues a task to fetch events if new block is Found
func blockScanTask(ctx context.Context, ethCl ethclient.Client, aqCl asynq.Client) error {
	currentBlock, err := ethCl.BlockNumber(ctx)
	var lastBlock uint64

	if err != nil {
		log.Errorf("BlockScan: %s", err)
		return err
	}
	if lastBlockVal := conf.RedisClient.Get(ctx, LastScannedBlockKey); lastBlockVal.Err() == redis.Nil {
		lastBlock = conf.Config.StartingBlockNumber
	} else {
		if r, parseErr := lastBlockVal.Int(); parseErr != nil {
			log.Errorf("blockScanTask: %s \nPossible issue is that somethings overwrote %s's value", parseErr, LastScannedBlockKey)
			return err
		} else {
			lastBlock = uint64(r)
		}
	}
	if lastBlock < currentBlock {
		for blockNum := lastBlock; blockNum < currentBlock; blockNum++ {
			_err := enqueueFetchBlockJob(aqCl, blockNum)
			if _err != nil {
				return _err
			}
		}
		status := conf.RedisClient.Set(ctx, LastScannedBlockKey, currentBlock, 0)
		if status != nil && status.Err() != nil {
			log.Errorf("BlockScan: %s", status.Err())
		}
	}
	return err
}

func enqueueUpdateUserBalJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(asynq.NewTask(UpdateUserBalanceKey, payload))
	return err
}

func enqueueParseBlockJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(BlockTask{BlockNumber: blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(asynq.NewTask(ParseBlockEventsKey, payload))
	return err
}

func enqueueFetchBlockJob(aqCl asynq.Client, blockNumber uint64) error {
	payload, err := json.Marshal(BlockTask{blockNumber})
	if err != nil {
		return err
	}
	_, err = aqCl.Enqueue(asynq.NewTask(FetchBlockEventsKey, payload))
	return err
}

// BlockEventsTask Fetches Block Events and stores them to mongo and enqueues another task for parsing them
func blockEventsTask(
	ctx context.Context,
	ethCl ethclient.Client,
	aqCl asynq.Client,
	monCl mongo.Collection,
	blockNum uint64,
) error {
	blockNumBigInt := big.NewInt(int64(blockNum))
	logs, err := ethCl.FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: blockNumBigInt,
			ToBlock:   blockNumBigInt,
		},
	)
	if err != nil {
		return err
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
	return err
}

func getLastBlock() (uint64, error) {
	var lastBlock uint64
	ctx := context.TODO()
	c, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	if res := conf.RedisClient.Get(c, LastScannedBlockKey); res.Err() != nil {
		return lastBlock, res.Err()
	} else {
		val, castErr := res.Uint64()
		lastBlock = val
		if castErr != nil {
			return lastBlock, fmt.Errorf("Can not cast to %d uint , err: %s", lastBlock, castErr)
		}
	}
	return lastBlock, nil
}
