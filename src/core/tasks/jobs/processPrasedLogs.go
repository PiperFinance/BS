package jobs

import (
	"context"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
)

// BPPL : Block Process Parse Log
type BPPL struct {
	bt        schema.BlockTask
	transfers []schema.LogTransfer
	// TODO: Add Other types here
}

func (l *BPPL) submit(c context.Context) error {
	return submitAllTransfers(c, l.bt, l.transfers)
}

// PrcoessParsedLogs takes parsed log interfaces and do actions needed for each on based on their types
func PrcoessParsedLogs(ctx context.Context, bt schema.BatchBlockTask, blockParsedLogs map[uint64][]interface{}) error {
	for blockNum := bt.FromBlockNum; blockNum <= bt.ToBlockNum; blockNum++ {
		bppl := BPPL{bt: schema.BlockTask{BlockNumber: blockNum, ChainId: bt.ChainId}}
		logs, ok := blockParsedLogs[blockNum]
		if !ok {
			conf.Logger.Warnw("ProccessParsedLogs: missing key in parsed map", "blockNum", blockNum)
			continue
		}
		for _, log := range logs {
			tr, ok := log.(schema.LogTransfer)
			if ok {
				bppl.transfers = append(bppl.transfers, tr)
			}
		}
		bppl.submit(ctx)
	}
	return nil
}
