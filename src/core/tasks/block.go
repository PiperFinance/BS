package tasks

import (
	"context"
	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

const (
	TypeBlockSearch = "block:search"
	// TODO - Add a model in db to save fechted block numbers
	// TODO - use gocache - not redis ...
	// RedisRetrivedBlocks
	LastScannedBlockKey = "block:lastScanned"
)

// BlockScanTask Enqueues a task to fetch events if new block is Found
func BlockScanTask(ctx context.Context, ethCl ethclient.Client, aqCl asynq.Client) error {
	currentBlock, err := ethCl.BlockNumber(ctx)
	var lastBlock uint64
	if err != nil {
		log.Errorf("BlockScan: %s", err)
	}
	lastBlock, _ = conf.RedisClient.Get(ctx, LastScannedBlockKey).Uint64()
	if lastBlock < currentBlock {
		for blockNum := uint64(0); blockNum < currentBlock; blockNum++ {
			EnqueueBlockJobs(blockNum)
		}
		status := conf.RedisClient.Set(ctx, LastScannedBlockKey, currentBlock, 0)
		if status != nil && status.Err() != nil {
			log.Errorf("BlockScan: %s", status.Err())
		}
	}
	return err
}

func EnqueueBlockJobs(blockNumber uint64) {
	
}
