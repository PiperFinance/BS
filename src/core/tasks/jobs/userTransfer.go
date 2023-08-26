package jobs

import (
	"context"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
	"github.com/ethereum/go-ethereum/common"
)

func UpdateUserBalJob(ctx context.Context, bt schema.BatchBlockTask, blockTransfers map[uint64][]schema.LogTransfer) error {
	for _, blockNum := range utils.SortedKeys[uint64, []schema.LogTransfer](blockTransfers) {
		_transfers, ok := blockTransfers[blockNum]
		if !ok {
			continue
		}
		if len(_transfers) > 0 {
			conf.Logger.Infof("processing [%d] block transfers", blockNum)
			thisBlock := schema.BlockTask{
				BlockNumber: blockNum,
				ChainId:     bt.ChainId,
			}
			if err := submitAllTransfers(ctx, thisBlock, _transfers); err != nil {
				return err
			}
		}
	}

	return nil
}

// findNewUsers users that are found for the first time
func findNewUsers(
	ctx context.Context,
	block schema.BlockTask,
	transfers []schema.LogTransfer,
) ([]contract_helpers.UserToken, error) {
	newUsers := make([]contract_helpers.UserToken, 0)
	// c ,cancel := context.WithCancel(ctx,)
	for _, transfer := range transfers {
		token := transfer.EmitterAddress

		for _, add := range []common.Address{transfer.From, transfer.To} {
			if err, yes := utils.IsNew(ctx, block.ChainId, add, token); err == nil && yes {
				// NOTE: check if there any duplication

				newUsers = append(newUsers, contract_helpers.UserToken{User: add, Token: token})
			} else if err != nil {
				return nil, err
			}
		}
	}
	return newUsers, nil
}
