package handlers

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hibiken/asynq"
)

type blockRange struct {
	startingBlock rpc.BlockNumber
	endingBlock   rpc.BlockNumber
}

func GetVacuumBlockRange(ctx context.Context) (blockRange, error) {
	// var lastScannedBlock uint64
	// br := blockRange{ startingBlock: LastVacuumedBlock,endingBlock: lsb - VacuumLogsHeight}

	// if res:=conf.RedisClient.Get(ctx,VacuumLogsKey);res.Err() !=nil{
	// 	if res.Err() == redis.Nil{
	// 		br.startingBlock
	// 		// Do nothing i guess
	// 	}
	// }
	// if res := conf.RedisClient.Get(ctx,LastScannedBlockKey);res.Err() == nil{
	// 	lsb ,err := res.Int64()
	// 	if err != nil{
	// 		fmt.Errorf("VacuumTask: %s",err)
	// 	}
	// 	// conf.RedisClient.GetOrSet(ctx,VacuumLogsKey,lsb)
	// 	_ = lsb
	// }
	return blockRange{}, nil
}

func VacuumLogHandler(ctx context.Context, task *asynq.Task) error {
	// Get LastVacuumed Block number

	// start int(getLastBlock() - conf.VacuumLogsHeight)
	// for i := ; i < ; i++ {

	// }getLastBlock()

	// // Save LastVacuumed BlockNumber
	return nil
}
