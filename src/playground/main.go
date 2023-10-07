package BSP

//         ╓                                                          ╖
//         ║                 Block Scanner Playground                 ║
//         ╙                                                          ╜

import (
	"context"
	"fmt"
	"log"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
)

func PlayHere() {
	parser := events.NewEventParser()
	_ = parser

	cl, rpc := conf.EthClientDebug(56)
	fmt.Printf("RPC: %s", rpc)
	block, err := cl.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	conf.Logger.Infof("Withdrawals: %+v \n", block.Withdrawals())
	conf.Logger.Infof("Body: %+v \n", block.Body())
	conf.Logger.Infof("Size: %+v \n", block.Size())
	conf.Logger.Infof("Trx: %+v \n", block.Transactions())
	for i, trx := range block.Transactions() {
		conf.Logger.Infof("Trx %d: %+v \n", i, trx)
	}
	conf.Logger.Infof("Block: %+v \n", block)
}
