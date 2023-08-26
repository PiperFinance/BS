package jobs

import (
	"context"
	"math"
	"math/big"
	"strings"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/core/types"
)

// SafeFilterLogs in case of "response size should not greater than " err happening we divide range
func SafeFilterLogs(
	ctx context.Context,
	bt schema.BatchBlockTask,
) ([]types.Log, error) {
	logs, err := conf.EthClient(bt.ChainId).FilterLogs(
		context.Background(),
		ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(bt.FromBlockNum)), // gte
			ToBlock:   big.NewInt(int64(bt.ToBlockNum)),   // lte
		},
	)
	if err == nil {
		return logs, err
	} else if bt.Range() > 1 &&
		strings.Contains(err.Error(), "response size should not greater than") {
		// NOTE: (s,e - math.ceil(r / 2)- 1) , (s+ math.floor(r / 2) ,e) -> ((1000, 1009), (1010, 1020))
		d := float64(bt.Range()) / 2
		c1_bt := schema.BatchBlockTask{
			FromBlockNum: bt.FromBlockNum,
			ToBlockNum:   bt.ToBlockNum - uint64(math.Ceil(d)+1),
			ChainId:      bt.ChainId,
		}
		c1, err := SafeFilterLogs(ctx, c1_bt)
		if err != nil {
			return nil, err
		}
		c2_bt := schema.BatchBlockTask{
			FromBlockNum: bt.FromBlockNum + uint64(math.Floor(d)),
			ToBlockNum:   bt.ToBlockNum,
			ChainId:      bt.ChainId,
		}
		c2, err := SafeFilterLogs(ctx, c2_bt)
		if err != nil {
			return nil, err
		}
		logs := append(c1, c2...)
		return logs, nil
	}
	return nil, err
}

// FetchBlockEvents Fetches All Block Events
func FetchBlockEvents(
	ctx context.Context,
	bt schema.BatchBlockTask,
) (error, map[uint64][]types.Log) {
	logs, err := SafeFilterLogs(ctx, bt)

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
