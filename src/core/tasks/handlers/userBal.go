package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// UpdateUserBalTaskHandler Updates Online User's Balance and then vacuums log record from database to save space
func UpdateUserBalTaskHandler(ctx context.Context, task *asynq.Task) error {
	// TODO - Why fixed timeout ?

	ctxFind, cancelFind := context.WithTimeout(ctx, 5*time.Minute)
	ctxDel, cancelDel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelFind()
	defer cancelDel()
	blockTask := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		// conf.Logger.Infof("Task ParseBlockEvents [%s] : Finished !", err)
		return err
	}

	// mutex := conf.RedisClient.ChainMutex(blockTask.ChainId, conf.UserBalanceRMutex)
	// defer mutex.Unlock()
	// if err := mutex.Lock(); err != nil {
	// 	conf.Logger.Warnf("UserBalHandler is Locked: %+v", blockTask)
	// 	return err
	// }

	cursor, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).Find(ctxFind, bson.M{
		"log.blockNumber": &blockTask.BlockNumber,
		"log.name":        events.TransferE,
	})
	defer cursor.Close(ctxFind)
	if err != nil {
		return err
	}
	transfers := make([]schema.LogTransfer, 0)
	transferIDs := make([]primitive.ObjectID, 0)
	for cursor.Next(ctx) {
		transfer := schema.LogTransfer{}
		if err := cursor.Decode(&transfer); err != nil {
			conf.Logger.Errorw("UserBal", "err", err, "block", blockTask)
			continue
		}
		transferIDs = append(transferIDs, transfer.ID)
		amount, ok := transfer.GetAmount()
		if ok && amount.Cmp(big.NewInt(0)) >= 1 {
			transfers = append(transfers, transfer)
		}
	}
	if len(transfers) > 0 {
		processTransferLogs(ctx, blockTask, transfers)
	}
	if len(transferIDs) > 0 {
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteMany(ctxDel, bson.M{"_id": bson.M{"$in": transferIDs}}); err != nil {
			return err
		}
	}

	bm := schema.BlockM{BlockNumber: blockTask.BlockNumber, ChainId: blockTask.ChainId}
	bm.SetAdded()
	if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
		ctx,
		bson.M{"no": blockTask.BlockNumber}, &bm); err != nil {
		return err
	}
	return err
}

func processTransferLogs(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) error {
	if err := updateTokens(ctx, block, transfers); err != nil {
		return err
	}
	err, newUsers := findNewRecords(ctx, block, transfers)
	if err != nil {
		return err
	}
	if err := updateUserTokens(ctx, block, newUsers); err != nil {
		return err
	}
	for _, trx := range transfers {
		if err := processTransferLog(ctx, block, trx); err != nil {
			conf.Logger.Errorw(err.Error(), "block", block.BlockNumber, "chain", block.ChainId)
		}
	}
	return nil
}

func updateTokens(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) error {
	col := conf.GetMongoCol(block.ChainId, conf.TokenColName)
	tokens := make([]interface{}, 0)
	for _, transfer := range transfers {
		if res := col.FindOne(ctx, bson.D{{Key: "_id", Value: transfer.EmitterAddress}}); res.Err() == mongo.ErrNoDocuments {
			tokens = append(tokens, bson.D{{Key: "_id", Value: transfer.EmitterAddress}})
		} else if res.Err() != nil {
			return res.Err()
		}
	}
	if len(tokens) > 1 {
		if _, err := col.InsertMany(ctx, tokens); err != nil {
			return err
		}
	}
	return nil
}

func isNew(ctx context.Context, chainId int64, user common.Address, token common.Address) (error, bool) {
	// TODO - Add user Limit here
	// FIXME - For Request CountReduction contracts contract and zero address is not included
	// if isLimited(ctx, chainId, user) {
	// 	return nil, false
	// }
	if !conf.OnlineUsers.IsAddressOnline(user) {
		return nil, false
	}

	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := userBalanceCol(chainId).FindOne(ctx, filter); res.Err() == mongo.ErrNoDocuments {
		return nil, true
	} else {

		if res.Err() == nil {
			conf.Logger.Infow("NewUserFinder", "user", user, "token", token, "err", res.Err())
		} else {
			conf.Logger.Errorw("NewUserFinder", "user", user, "token", token, "err", res.Err())
		}
		return res.Err(), false
	}
}

func isLimited(ctx context.Context, chainId int64, user common.Address) bool {
	return isAddressNull(ctx, chainId, user) || isAddressToken(ctx, chainId, user) || isUserBanned(ctx, chainId, user)
}

func isUserBanned(ctx context.Context, chainId int64, user common.Address) bool {
	if res := conf.GetMongoCol(chainId, conf.UserBalColName).FindOne(ctx, bson.D{{Key: "_id", Value: user.String()}}); res.Err() == mongo.ErrNoDocuments {
		return false
	} else if res.Err() == nil {
		return true
	} else {
		conf.Logger.Errorw("BannedUsers", "err", res.Err(), "user", user, "chain", chainId)
		return false
	}
}

func isAddressNull(ctx context.Context, chainId int64, user common.Address) bool {
	return user.Big().Cmp(big.NewInt(0)) < 1
}

func isAddressToken(ctx context.Context, chainId int64, user common.Address) bool {
	if res := conf.GetMongoCol(chainId, conf.TokenColName).FindOne(ctx, bson.D{{Key: "_id", Value: user}}); res.Err() == mongo.ErrNoDocuments {
		return false
	} else if res.Err() == nil {
		return true
	} else {
		conf.Logger.Errorw("IsAddressAToken", "err", res.Err(), "user", user, "chain", chainId)
		return false
	}
}

func updateUserTokens(ctx context.Context, blockTask schema.BlockTask, usersTokens []contract_helpers.UserToken) error {
	if len(usersTokens) < 1 {
		return nil
	}
	conf.NewUsersCount.AddFor(blockTask.ChainId, uint64(len(usersTokens)))
	conf.MultiCallCount.Add(blockTask.ChainId)
	// bal := contract_helpers.EasyBalanceOf{UserTokens: usersTokens, ChainId: blockTask.ChainId, BlockNumber: int64(blockTask.BlockNumber)}
	bal := contract_helpers.EasyBalanceOf{UserTokens: usersTokens, ChainId: blockTask.ChainId, BlockNumber: int64(blockTask.BlockNumber) - 1}
	if err := bal.Execute(ctx); err != nil {
		// conf.Logger.Error(err)
		return err
	}
	col := userBalanceCol(blockTask.ChainId)
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
			TrxCount:  0,
			ChangedAt: blockTask.BlockNumber,
			StartedAt: blockTask.BlockNumber,
			Balance:   userToken.Balance.String(),
		})
	}
	if len(balances) > 0 {
		if res, err := col.InsertMany(ctx, balances); err != nil {
			return err
		} else {
			conf.Logger.Info(res)
		}
	}
	return nil
}

func findNewRecords(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) (error, []contract_helpers.UserToken) {
	newUsers := make([]contract_helpers.UserToken, 0)
	for _, transfer := range transfers {
		token := transfer.EmitterAddress
		if err, yes := isNew(ctx, block.ChainId, transfer.From, token); err == nil && yes {
			newUsers = append(newUsers, contract_helpers.UserToken{User: transfer.From, Token: token})
		} else if err != nil {
			return err, nil
		}
	}
	return nil, newUsers
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

func processUserBal(ctx context.Context, blockTask schema.BlockTask, user common.Address, token common.Address, amount *big.Int) (*schema.UserBalance, error) {
	userBal := schema.UserBalance{
		User:      user,
		Token:     token,
		ChangedAt: blockTask.BlockNumber,
	}
	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := userBalanceCol(blockTask.ChainId).FindOne(ctx, filter); res.Err() == mongo.ErrNoDocuments {
		// NOTE - Record might have been ignored
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
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "bal", Value: userBal.GetBalanceStr()}, {Key: "c_t", Value: blockTask.BlockNumber}, {Key: "count", Value: userBal.TrxCount + 1}}}}
	_, err := userBalanceCol(blockTask.ChainId).UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &userBal, nil
}
