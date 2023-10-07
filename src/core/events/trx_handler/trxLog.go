package trx_handler

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

var submitionLock = sync.Mutex{}

func (h *UserTrxHandler) submitAllTransfers(ctx context.Context, chainId int64, blockNumber uint64, transfers []*schema.LogTransfer) error {
	// NOTE: Store new token's in mongoDB
	if err := h.updateTokens(ctx, chainId, transfers); err != nil {
		return err
	}

	// NOTE: find new users
	newUsers, err := h.findNewUsers(ctx, chainId, transfers)
	if err != nil {
		return err
	}

	// NOTE: Chunk Call new users
	wg := sync.WaitGroup{}
	for _, chunk := range utils.ChunkNewUserCalls(chainId, newUsers) {
		wg.Add(1)
		go func(_chunk []contract_helpers.UserToken) {
			if _err := h.updateUserTokens(ctx, chainId, blockNumber, _chunk); err != nil {
				err = _err
			}
			wg.Done()
		}(chunk)
	}
	wg.Wait()
	if err != nil {
		// NOTE: even if one of the results respond with err the whole task will be retried
		return err
	}

	// NOTE:  submits transfers in userbalance collection
	for i, trx := range transfers {
		submitionLock.Lock()
		if err := sumbitTransfer(ctx, chainId, trx.BlockNumber, uint64(i), trx); err != nil {
			conf.Logger.Errorw(err.Error(), "block", trx.BlockNumber, "trxindex", trx.LogIndex, "i", i, "chain", chainId)
		}
		submitionLock.Unlock()
	}

	// NOTE: store transfer maybe in db
	if conf.Config.SaveAllTransferLogs {
		for _, trx := range transfers {
			if _, err := conf.GetMongoCol(chainId, conf.TransfersColName).InsertOne(ctx, trx); err != nil {
				return err
			}
		}
	}

	return nil
}

// sumbitTransfer increase to address of transfer and subtracts amount from
func sumbitTransfer(ctx context.Context, chainId int64, blockNumber uint64, iterationIdx uint64, transfer *schema.LogTransfer) error {
	var amount *big.Int
	if b, ok := transfer.GetAmount(); ok {
		amount = b
	} else {
		return fmt.Errorf("transfer log get amount failure, transfer=%+v", transfer)
	}

	if err := conf.RedisClient.ReentrancyCheckSet(ctx, chainId, fmt.Sprintf("%d-%d", transfer.BlockNumber, transfer.LogIndex), iterationIdx); err != nil {
		return err
	}

	if _, err := processUserBal(
		ctx, chainId, blockNumber,
		transfer.To, transfer.EmitterAddress,
		amount); err != nil {
		return err
	}

	if _, err := processUserBal(
		ctx, chainId, blockNumber,
		transfer.From, transfer.EmitterAddress,
		amount.Neg(amount)); err != nil {
		return err
	}
	return nil
}

// processUserBal executes update query in db
func processUserBal(ctx context.Context, chainId int64, blockNumber uint64, user common.Address, token common.Address, amount *big.Int) (*schema.UserBalance, error) {
	userBal := schema.UserBalance{
		User:      user,
		Token:     token,
		ChangedAt: blockNumber,
	}
	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := conf.GetMongoCol(chainId, conf.UserBalColName).FindOne(ctx, filter); res.Err() == mongo.ErrNoDocuments {
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
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "bal", Value: userBal.GetBalanceStr()}, {Key: "c_t", Value: blockNumber}, {Key: "count", Value: userBal.TrxCount + 1}}}}

	// TODO: - Make this Update Many
	_, err := conf.GetMongoCol(chainId, conf.UserBalColName).UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &userBal, nil
}
