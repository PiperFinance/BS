package test_core

import (
	"context"
	"math/big"
	"testing"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/ethereum/go-ethereum"
)

func NetworkCapacities(t *testing.T) {
	// 250 -> 25 having 2640
	results := make(map[int64]uint64)
	for _, chain := range conf.Config.SupportedChains {

		cb, _ := conf.EthClient(chain).BlockNumber(context.Background())
		var i uint64
		for {
			i++
			cl, rpc := conf.EthClientDebug(chain)
			logs, err := cl.FilterLogs(
				context.Background(),
				ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(cb - i)),
					ToBlock:   big.NewInt(int64(cb)),
				},
			)
			if err != nil {
				conf.Logger.Error(err)
				results[chain] = i
				break
			}
			conf.Logger.Infow("res", "rpc", rpc, "len", len(logs), "length", i, "query", []uint64{cb, cb - i})
		}
	}
	// 1: 1 = 0x1
	// 250: 66 = 0x42
	// 56: 19 = 0x13
	// 137: 15 = 0xf
	// 42161: 313 = 0x139
	// 9001: 87 = 0x57
	// 58: 2698 = 0xa8a
	// 43114: 55 = 0x37
	// 100: 228 = 0xe4
	// 2021: 1 = 0x1
	// 1284: 59 = 0x3b
	conf.Logger.Info(results)
}
