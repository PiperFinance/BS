package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/enqueuer"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

// FetchBlockTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func FetchBlockTaskHandler(ctx context.Context, task *asynq.Task) error {
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

func checkResults(ctx context.Context, b schema.BatchBlockTask) {
	// NOTE: DEBUG
	col := userBalanceCol(b.ChainId)

	curs, err := col.Find(ctx, bson.M{})
	if err != nil {
		conf.Logger.Error(err)
	} else {
		for curs.Next(ctx) {
			val := schema.UserBalance{}
			curs.Decode(&val)
			bal, _ := val.GetBalance()
			if bal.Cmp(big.NewInt(0)) == -1 {
				conf.RedisClient.IncrHSet(ctx, fmt.Sprintf("BS:NVR:%d", 56), val.TokenStr)
				if val.TokenStr == "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c" {
					fmt.Println("NEGATIVE BAL")
				}
			}
		}
	}
}

// fetchBlockEventsJob Fetches Block Events
func fetchBlockEventsJob(
	ctx context.Context,
	bt schema.BatchBlockTask,
) (error, map[uint64][]types.Log) {
	// TODO: Retry With reduced range is this fails
	logs, err := conf.EthClient(bt.ChainId).FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(bt.FromBlockNum)), // gte
			ToBlock:   big.NewInt(int64(bt.ToBlockNum)),   // lte
		},
	)
	// NOTE: DEUBG
	conf.CallCount.Add(bt.ChainId)
	if err != nil {
		conf.FailedCallCount.Add(bt.ChainId)
		return &utils.RpcError{Err: err, ChainId: bt.ChainId, ToBlockNumber: bt.ToBlockNum, FromBlockNumber: bt.FromBlockNum, Name: "BlockFetch"}, nil
	}

	if len(logs) < 1 {
		return nil, nil
	}
	for i := bt.FromBlockNum; i <= bt.ToBlockNum; i++ {
		conf.Logger.Infow("Fetched", "block", i)
	}
	res := make(map[uint64][]types.Log, bt.ToBlockNum-bt.FromBlockNum+1)
	for i := bt.FromBlockNum; i <= bt.ToBlockNum; i++ {
		res[i] = make([]types.Log, 0)
	}
	for _, _log := range logs {
		if !_log.Removed {
			res[_log.BlockNumber] = append(res[_log.BlockNumber], _log)
		}
	}

	return err, res
}
