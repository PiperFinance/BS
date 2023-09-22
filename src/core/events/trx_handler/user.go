package trx_handler

import (
	"context"

	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
	"github.com/ethereum/go-ethereum/common"
)

// findNewUsers users that are found for the first time
func (h *UserTrxHandler) findNewUsers(
	ctx context.Context,
	chainId int64,
	transfers []*schema.LogTransfer,
) ([]contract_helpers.UserToken, error) {
	newUsers := make([]contract_helpers.UserToken, 0)
	for _, transfer := range transfers {
		token := transfer.EmitterAddress

		for _, add := range []common.Address{transfer.From, transfer.To} {
			if err, yes := utils.IsNew(ctx, chainId, add, token); err == nil && yes {
				// NOTE: check if there any duplication

				newUsers = append(newUsers, contract_helpers.UserToken{User: add, Token: token})
			} else if err != nil {
				return nil, err
			}
		}
	}
	return newUsers, nil
}
