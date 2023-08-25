package jobs

import (
	"context"
	"math/big"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// FetchBlockEvents Fetches All Block Events
func FetchBlockEvents(
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
