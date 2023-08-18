package handlers

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/hibiken/asynq"
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

// saveBlocks insert newly scanned blocks into db ( with unique )
func saveBlocks(ctx context.Context, chain int64, from, to uint64) error {
	newBlocks := make([]interface{}, 0)
	for i := from; i <= to; i++ {
		b := schema.BlockM{BlockNumber: i, ChainId: chain}
		b.SetScanned()
		newBlocks = append(newBlocks, b)
	}
	_, err := conf.GetMongoCol(chain, conf.BlockColName).InsertMany(ctx, newBlocks)
	return err
}

// scanBlockJob Enqueues a task to fetch events if new block range is met the  height
// from head up current block - delay
// 🚧 inner app queries are gte -> lte
func scanBlockJob(ctx context.Context, blockTask schema.BatchBlockTask, aqCl asynq.Client) error {
	var head uint64
	chain := blockTask.ChainId
	batchSize := conf.BatchLogMaxHeight(chain)
	currentBlock, err := conf.LatestBlock(ctx, blockTask.ChainId)
	if err != nil {
		return &utils.RpcError{Err: err, ChainId: blockTask.ChainId, ToBlockNumber: blockTask.ToBlockNum, FromBlockNumber: blockTask.FromBlockNum, Name: "BlockScan"}
	}

	if cmd := conf.RedisClient.Get(ctx, tasks.LastScannedBlockKey(chain)); cmd.Err() != nil {
		return cmd.Err()
	} else {
		if r, parseErr := cmd.Int(); parseErr != nil {
			conf.Logger.Errorf("blockScanTask: %s \nPossible issue is that somethings overwrote %s's value", parseErr, tasks.LastScannedBlockKey(chain))
			return err
		} else {
			head = uint64(r)
		}
	}
	/*
		NOTE:
			Head : is Last Scanned Block
			currentBlock : is Head of Network in called by web3
			batch size : is dynamic size per chain
			h = 100 	bs = 5 		cb = 116
			h + bs = 105 > 106 [OK]
			cb - h (16)
		 	 --------      = 3.1 ~= 2
		    	bs (5)
			(100-104) (105-109) (110-114)
	*/

	if head+batchSize < currentBlock {
		batchCount := (currentBlock - head) / batchSize
		newHead := head
		conf.Logger.Infow("BlockScan", "block", currentBlock, "head", head, "b-size", batchSize, "b-count", batchCount)

		for j := uint64(0); j < batchCount; j++ {
			b := schema.BatchBlockTask{
				FromBlockNum: head + (j * batchSize),
				ToBlockNum:   head + ((j + 1) * batchSize) - 1,
				ChainId:      chain,
			}
			if b.ToBlockNum > currentBlock {
				b.ToBlockNum = currentBlock
			}
			if b.FromBlockNum == b.ToBlockNum {
				break
			}
			if err := saveBlocks(ctx, chain, b.FromBlockNum, b.ToBlockNum); err != nil {
				return err
			}
			if err := enqueuer.EnqueueProcessBlockJob(aqCl, b); err != nil {
				return err
			}
			for i := b.FromBlockNum; i <= b.ToBlockNum; i++ {
				conf.Logger.Infow("Enqueue Scan", "block", i)
			}
			newHead = b.ToBlockNum + 1
			// NOTE: DEBUG
			conf.NewBlockCount.AddFor(chain, uint64(b.ToBlockNum-b.FromBlockNum))
		}

		if cmd := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), newHead, 0); cmd.Err() != nil {
			return cmd.Err()
		}
	}
	return err
}
