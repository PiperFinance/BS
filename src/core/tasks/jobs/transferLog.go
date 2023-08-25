package jobs

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func submitAllTransfers(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) error {
	if err := updateTokens(ctx, block, transfers); err != nil {
		return err
	}
	newUsers, err := findNewUsers(ctx, block, transfers)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, chunk := range utils.ChunkNewUserCalls(block.ChainId, newUsers) {
		wg.Add(1)
		go func(_chunk []contract_helpers.UserToken) {
			if _err := updateUserTokens(ctx, block, _chunk); err != nil {
				err = _err
			}
			wg.Done()
		}(chunk)
	}
	wg.Wait()
	if err != nil {
		return err
	}
	for _, trx := range transfers {
		if err := sumbitTransfer(ctx, block, trx); err != nil {
			conf.Logger.Errorw(err.Error(), "block", block.BlockNumber, "chain", block.ChainId)
		}
	}

	return nil
}

func sumbitTransfer(ctx context.Context, block schema.BlockTask, transfer schema.LogTransfer) error {
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
