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
		return &utils.RpcError{Err: err, ChainId: blockTask.ChainId, ToBlockNumber: blockTask.ToBlockNumber, FromBlockNumber: blockTask.FromBlockNumber, Name: "BlockScan"}
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
	// Head : is Last Scanned Block
	// currentBlock : is Head of Network in called by web3
	// batch size : is dynamic size per chain

	// h = 100 	bs = 5 		cb = 116
	// h + bs = 105 > 106 [OK]
	// cb - h = 16
	//  --------      = 3.1 ~= 2
	//     bs (5)
	// (100-104) (105-109) (110-114)

	if head+batchSize < currentBlock {
		batchCount := (currentBlock - head) / batchSize
		conf.Logger.Infow("BlockScan", "block", currentBlock, "head", head, "b-size", batchSize, "b-count", batchCount)

		newHead := head

		for j := uint64(0); j < batchCount; j++ {
			b := schema.BatchBlockTask{
				FromBlockNumber: head + (j * batchSize),
				ToBlockNumber:   head + ((j + 1) * batchSize) - 1,
				ChainId:         chain,
			}
			if b.ToBlockNumber > currentBlock {
				b.ToBlockNumber = currentBlock
			}
			if b.FromBlockNumber == b.ToBlockNumber {
				break
			}
			if err := saveBlocks(ctx, chain, b.FromBlockNumber, b.ToBlockNumber); err != nil {
				return err
			}
			if _err := enqueuer.EnqueueFetchBlockJob(aqCl, b); _err != nil {
				return _err
			}
			for i := b.FromBlockNumber; i <= b.ToBlockNumber; i++ {
				conf.Logger.Infow("Enqueue Scan", "block", i)
			}
			newHead = b.ToBlockNumber + 1
			conf.NewBlockCount.AddFor(chain, uint64(b.ToBlockNumber-b.FromBlockNumber))
		}

		if cmd := conf.RedisClient.Set(ctx, tasks.LastScannedBlockKey(chain), newHead, 0); cmd.Err() != nil {
			return cmd.Err()
		}
	}
	return err
}
