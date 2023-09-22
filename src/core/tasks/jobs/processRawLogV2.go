package jobs

import (
	"context"

	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/core/types"
)

// ProccessLogs  Parses raw logs and adds Do needed actions predefined
func ProccessLogs(ctx context.Context, bt schema.BatchBlockTask, blocksLogs map[uint64][]types.Log) error {
	for blockNo, logs := range blocksLogs {
		parser := events.NewEventParser()
		for _, log := range logs {
			err := parser.Parse(log)
			if err != nil {
				// TODO: Add Some sort of handler to retry on blocks that are actually parsed
				return err
			}
		}
		if err := parser.Flush(ctx, bt.ChainId, blockNo); err != nil {
			return err
		}
	}
	return nil
}
