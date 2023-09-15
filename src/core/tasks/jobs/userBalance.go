package jobs

import (
	"context"
	"math/big"
	"sync"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
)

var insertUpdates sync.Mutex

func init() {
	insertUpdates = sync.Mutex{}
}

// updateUserTokens
// - uses multicall to update user bal
// - store user - token on both mongo and redis
func updateUserTokens(ctx context.Context, bt schema.BlockTask, usersTokens []contract_helpers.UserToken) error {
	if len(usersTokens) < 1 {
		return nil
	}
	conf.NewUsersCount.AddFor(bt.ChainId, uint64(len(usersTokens)))
	conf.MultiCallCount.Add(bt.ChainId)
	bal := contract_helpers.EasyBalanceOf{UserTokens: usersTokens, ChainId: bt.ChainId, BlockNumber: int64(bt.BlockNumber) - 1}
	if err := bal.Execute(ctx); err != nil {
		return err
	}
	col := conf.GetMongoCol(bt.ChainId, conf.UserBalColName)
	balances := make([]interface{}, 0)

	for _, userToken := range bal.UserTokens {
		if userToken.Balance == nil {
			conf.Logger.Errorf("token:%s user:%d %+v", userToken.User.String(), userToken.Token.String(), userToken)
			continue
		}
		balances = append(balances, schema.UserBalance{
			User:      userToken.User,
			Token:     userToken.Token,
			UserStr:   userToken.User.String(),
			TokenStr:  userToken.Token.String(),
			TokenId:   conf.FindTokenId(bal.ChainId, userToken.Token),
			TrxCount:  0,
			ChangedAt: bt.BlockNumber,
			StartedAt: bt.BlockNumber,
			Balance:   userToken.Balance.String(),
		})
		if userToken.Balance.Cmp(big.NewInt(0)) == -1 {
			conf.Logger.Errorf("Negative Balance %v", userToken)
		}
	}
	if len(balances) > 0 {
		// NOTE: DEBUG - After running this shows no sign of a negative value
		insertUpdates.Lock()
		defer insertUpdates.Unlock()
		if _, err := col.InsertMany(ctx, balances); err != nil {
			return err
		}
	}
	return nil
}
