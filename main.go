package main

import (
	"context"
	"fmt"
	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/ethereum/go-ethereum"
	"github.com/kamva/mgm/v3"
	"log"
	"math/big"
)

// ONLY FOR TESTING PURPOSES ...
func main() {
	current, _err := conf.EthClient.BlockNumber(context.Background())
	if _err != nil {
		log.Fatal(_err)
	}
	fmt.Printf("?/%d\r", current)
	for i := int64(current) - 100; i < int64(current); i++ {
		fmt.Printf("[%d/%d] remains:%d\r", i, current, (int64(current) - i))
		fromBlock := big.NewInt(i)
		toBlock := big.NewInt(i)
		logs, err := conf.EthClient.FilterLogs(
			context.Background(),
			ethereum.FilterQuery{
				FromBlock: fromBlock,
				ToBlock:   toBlock},
		)
		for _, log := range logs {

			mgm.Coll(&LogColl{}).CreateWithCtx(context.Background(), &LogColl{
				Address:     log.Address,
				Data:        log.Data,
				Index:       log.Index,
				Topics:      log.Topics,
				TxIndex:     log.TxIndex,
				BlockNumber: log.BlockNumber,
				BlockHash:   log.BlockHash,
				Removed:     log.Removed,
				TxHash:      log.TxHash,
			}, nil)
		}
		if err != nil {
			fmt.Println(err)
		}
	}
}
