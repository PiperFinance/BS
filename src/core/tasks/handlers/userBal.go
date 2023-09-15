package handlers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
)

func updateUserTokens(ctx context.Context, bt schema.BlockTask, usersTokens []contract_helpers.UserToken) error {
	if len(usersTokens) < 1 {
		return nil
	}
	conf.NewUsersCount.AddFor(bt.ChainId, uint64(len(usersTokens)))
	conf.MultiCallCount.Add(bt.ChainId)
	// TODO: chunk batch calls !
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
		if _, err := col.InsertMany(ctx, balances); err != nil {
			return err
		}
	}
	return nil
}

func processTransferLog(ctx context.Context, block schema.BlockTask, transfer schema.LogTransfer) error {
	var amount *big.Int
	if b, ok := transfer.GetAmount(); ok {
		amount = b
	} else {
		return fmt.Errorf("transfer log get amount failure, transfer=%+v", transfer)
	}

	if _, err := processUserBal(
		ctx, block,
		transfer.To, transfer.EmitterAddress,
		amount); err != nil {
		return err
	}

	if _, err := processUserBal(
		ctx, block,
		transfer.From, transfer.EmitterAddress,
		amount.Neg(amount)); err != nil {
		return err
	}
	return nil
}

func processUserBal(ctx context.Context, bt schema.BlockTask, user common.Address, token common.Address, amount *big.Int) (*schema.UserBalance, error) {
	userBal := schema.UserBalance{
		User:      user,
		Token:     token,
		ChangedAt: bt.BlockNumber,
	}
	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := conf.GetMongoCol(bt.ChainId, conf.UserBalColName).FindOne(ctx, filter); res.Err() == mongo.ErrNoDocuments {
		// NOTE:  Record might have been ignored
		return nil, nil
	} else if res.Err() != nil {
		return nil, res.Err()
	} else {
		if err := res.Decode(&userBal); err != nil {
			return nil, err
		}
	}
	if err := userBal.AddBal(amount); err != nil {
		return nil, err
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "bal", Value: userBal.GetBalanceStr()}, {Key: "c_t", Value: bt.BlockNumber}, {Key: "count", Value: userBal.TrxCount + 1}}}}

	// TODO: - Make this Update Many
	_, err := conf.GetMongoCol(bt.ChainId, conf.UserBalColName).UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &userBal, nil
}
