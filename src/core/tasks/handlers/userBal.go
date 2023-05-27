package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/contracts"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func lock(key string) {
}

func userBalanceCol(chain int64) *mongo.Collection {
	return conf.GetMongoCol(chain, conf.UserBalColName)
}

func TokenVolumeCol(chain int64) *mongo.Collection {
	return conf.GetMongoCol(chain, conf.TokenVolumeColName)
}

func getBalance(ctx context.Context, block schema.BlockTask, user common.Address, token common.Address) (*big.Int, error) {
	if caller, err := contracts.NewERC20Caller(token, conf.EthClient(block.ChainId)); err != nil {
		return nil, err
	} else {
		return caller.BalanceOf(&bind.CallOpts{
			Context: ctx, BlockNumber: big.NewInt(int64(block.BlockNumber - 1)),
		}, user)
	}
}

func ChainMutex(i int64) {
}

// UpdateUserBalTaskHandler Updates Online User's Balance and then vacuums log record from database to save space
func UpdateUserBalTaskHandler(ctx context.Context, task *asynq.Task) error {
	// TODO - Why fixed timeout ?

	ctxFind, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	blockTask := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		// conf.Logger.Infof("Task ParseBlockEvents [%s] : Finished !", err)
		return err
	}

	mutex := conf.RedisClient.ChainMutex(blockTask.ChainId, conf.UserBalanceRMutex)
	defer mutex.Unlock()
	if err := mutex.Lock(); err != nil {
		conf.Logger.Warnf("UserBalHandler is Locked: %+v", blockTask)
		return err
	}
	cursor, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).Find(ctxFind, bson.M{
		"log.blockNumber": &blockTask.BlockNumber,
		// TODO - Chain check here ....
		"log.name": events.TransferE,
	})
	defer cursor.Close(ctxFind)
	if err != nil {
		return err
	}
	transfers := make([]schema.LogTransfer, 0)
	for cursor.Next(ctx) {
		transfer := schema.LogTransfer{}
		if err := cursor.Decode(&transfer); err != nil {
			conf.Logger.Error(err)
			continue
		}
		transfers = append(transfers, transfer)
	}

	processTransferLogs(ctx, blockTask, transfers)
	// flushTransferLogs(ctx, blockTask, transfers)
	// if _, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteOne(ctxFind, bson.M{"_id": transfer.ID}); err != nil {
	// 	conf.Logger.Error(err)
	// } else {
	// 	// conf.Logger.Info(res)
	// }

	// if res, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteMany(ctxFind, bson.M{
	// 	"log.blockNumber": &block.BlockNumber,
	// 	"log.name":        events.TransferE,
	// }); err != nil {
	// 	conf.Logger.Errorf("BlockEventsTaskHandler")
	// } else {
	// 	conf.Logger.Infof("Deleted Logs : %s Deleted", res.DeletedCount)
	// }
	bm := schema.BlockM{BlockNumber: blockTask.BlockNumber}
	bm.SetParsed()
	if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
		ctx,
		bson.M{"no": blockTask.BlockNumber}, &bm); err != nil {
		conf.Logger.Errorf("BlockEventsTaskHandler")
	} else {
		// conf.Logger.Infof("Replace Result : %s modified", res.ModifiedCount)
	}
	if err != nil {
		conf.Logger.Errorf("Task UpdateUserBal [%d] : Err : %s !", blockTask.BlockNumber, err)
	} else {
		// conf.Logger.Infof("Task UpdateUserBal [%d] : Finished !", block.BlockNumber)
	}
	return err
}

func processTransferLogs(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) error {
	// NOTE - Make sure this action blocks the this chain work flow [ No new block event fetch ...]
	// STUB - used redis redsync lock

	// NOTE - Find user's with no previous balance
	// findNewRecordsUnsafe()
	// // NOTE - Update user's old balance
	// updateUserTokens()
	// // NOTE - Process Transfers as usual
	// processTransferLog()
	// NOTE - unlock task log
	// NOTE - Flush logs

	return nil
}

func isNew(ctx context.Context, chainId int64, user common.Address, token common.Address) bool {
	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := userBalanceCol(chainId).FindOne(ctx, filter); res.Err() == mongo.ErrNoDocuments {
		return true
	} else {
		conf.Logger.Warnf("NewUserFinder: user=%s ,token=%s , err=%+v", user.String(), token.String(), res.Err())
		return false
	}
}

func updateUserTokens(ctx context.Context, blockTask schema.BlockTask, usersTokens []contract_helpers.UserToken) error {
	// TODO - multicall here
	return nil
}

func findNewRecordsUnsafe(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) []contract_helpers.UserToken {
	newUsers := make([]contract_helpers.UserToken, 0)
	for _, transfer := range transfers {
		token := transfer.EmitterAddress
		if isNew(ctx, block.ChainId, transfer.From, token) {
			newUsers = append(newUsers, contract_helpers.UserToken{User: transfer.From, Token: token})
		}
	}
	return newUsers
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
		transfer.From, transfer.EmitterAddress,
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

func processUserBal(ctx context.Context, blockTask schema.BlockTask, user common.Address, token common.Address, amount *big.Int) (*schema.UserBalance, error) {
	userBal := schema.UserBalance{
		User:      user,
		Token:     token,
		ChangedAt: blockTask.BlockNumber,
	}
	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := userBalanceCol(blockTask.ChainId).FindOne(ctx, filter); res.Err() != nil {
		return nil, res.Err()
	} else {
		if err := res.Decode(&userBal); err != nil {
			return nil, err
		}
	}
	if err := userBal.AddBal(amount); err != nil {
		return nil, err
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "bal", Value: userBal.GetBalanceStr()}, {Key: "c_t", Value: blockTask.BlockNumber}}}}
	_, err := userBalanceCol(blockTask.ChainId).UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	// conf.Logger.Infof("Modified: %d , MatchedCount: %d ", res.ModifiedCount, res.MatchedCount)
	return &userBal, nil
}
