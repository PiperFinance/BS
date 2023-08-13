package handlers

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
)

func userBalanceCol(chain int64) *mongo.Collection {
	return conf.GetMongoCol(chain, conf.UserBalColName)
}

func TokenVolumeCol(chain int64) *mongo.Collection {
	return conf.GetMongoCol(chain, conf.TokenVolumeColName)
}

func chunkNewUserCalls(chain int64, users []contract_helpers.UserToken) [][]contract_helpers.UserToken {
	batchSize := int(conf.MulticallMaxSize(chain))
	chunkCount := (len(users) / batchSize) + 1
	r := make([][]contract_helpers.UserToken, chunkCount)
	for i := 0; i < chunkCount; i++ {
		startingIndex := i * batchSize
		endingIndex := (i + 1) * batchSize
		if endingIndex > len(users) {
			endingIndex = len(users)
		}
		r[i] = users[startingIndex:endingIndex]
	}
	return r
}

func processTransferLogs(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) error {
	if err := updateTokens(ctx, block, transfers); err != nil {
		return err
	}
	newUsers, err := findNewUsers(ctx, block, transfers)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, chunk := range chunkNewUserCalls(block.ChainId, newUsers) {
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
		if err := processTransferLog(ctx, block, trx); err != nil {
			conf.Logger.Errorw(err.Error(), "block", block.BlockNumber, "chain", block.ChainId)
		}
	}

	// NOTE: DEBUG
	col := userBalanceCol(block.ChainId)
	for _, trx := range transfers {
		curs, err := col.Find(ctx, bson.M{"user": trx.From, "token": trx.EmitterAddress})
		if err != nil {
			conf.Logger.Error(err)
		} else {
			for curs.Next(ctx) {
				val := schema.UserBalance{}
				curs.Decode(&val)
				bal, _ := val.GetBalance()
				if bal.Cmp(big.NewInt(0)) == -1 {
					conf.RedisClient.IncrHSet(ctx, fmt.Sprintf("BS:NVR:%d", 56), val.TokenStr)
					if val.TokenStr == "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c" {
						conf.Logger.Errorw("Negative Value Record", "trx", trx, "bal", bal.String(), "rec", val, "ObjId", trx.ID)
					}
				}
			}
		}
		curs, err = col.Find(ctx, bson.M{"user": trx.To, "token": trx.EmitterAddress})
		if err != nil {
			conf.Logger.Error(err)
		} else {
			for curs.Next(ctx) {
				val := schema.UserBalance{}
				curs.Decode(&val)
				bal, _ := val.GetBalance()
				if bal.Cmp(big.NewInt(0)) == -1 {
					conf.RedisClient.IncrHSet(ctx, fmt.Sprintf("BS:NVR:%d", 56), val.TokenStr)
					if val.TokenStr == "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c" {
						conf.Logger.Errorw("Negative Value Record", "trx", trx, "bal", bal.String(), "rec", val, "ObjId", trx.ID)
					}
				}
			}
		}
	}

	return nil
}

func updateTokens(ctx context.Context, block schema.BlockTask, transfers []schema.LogTransfer) error {
	col := conf.GetMongoCol(block.ChainId, conf.TokenColName)
	// tokens := make([]interface{}, 0)
	uniqueTokens := make([]common.Address, 0)
	var tokenExists bool
	for _, transfer := range transfers {
		_token := transfer.EmitterAddress
		tokenExists = true
		for _, token := range uniqueTokens {
			if token == _token {
				tokenExists = false
				break
			}
		}
		if tokenExists {
			uniqueTokens = append(uniqueTokens, _token)
		}
	}
	for _, token := range uniqueTokens {
		if count, err := col.CountDocuments(ctx, bson.D{{Key: "_id", Value: token}}); count == 0 || err == mongo.ErrNoDocuments {
			// tokens = append(tokens, )
			// TODO - check err later
			col.InsertOne(ctx, bson.D{{Key: "_id", Value: token}})
		} else if err != nil {
			return err
		}
	}
	return nil
}

func updateUserTokens(ctx context.Context, blockTask schema.BlockTask, usersTokens []contract_helpers.UserToken) error {
	if len(usersTokens) < 1 {
		return nil
	}
	conf.NewUsersCount.AddFor(blockTask.ChainId, uint64(len(usersTokens)))
	conf.MultiCallCount.Add(blockTask.ChainId)
	// TODO: chunk batch calls !
	bal := contract_helpers.EasyBalanceOf{UserTokens: usersTokens, ChainId: blockTask.ChainId, BlockNumber: int64(blockTask.BlockNumber) - 1}
	if err := bal.Execute(ctx); err != nil {
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
			TokenId:   conf.FindTokenId(bal.ChainId, userToken.Token),
			TrxCount:  0,
			ChangedAt: blockTask.BlockNumber,
			StartedAt: blockTask.BlockNumber,
			Balance:   userToken.Balance.String(),
		})
		if userToken.Balance.Cmp(big.NewInt(0)) == -1 {
			conf.Logger.Errorf("Negative Balance %v", userToken)
		}
	}
	if len(balances) > 0 {
		// DEBUG - After running this shows no sign of a negative value
		if _, err := col.InsertMany(ctx, balances); err != nil {
			return err
		}
	}
	return nil
}

func findNewUsers(
	ctx context.Context,
	block schema.BlockTask,
	transfers []schema.LogTransfer,
) ([]contract_helpers.UserToken, error) {
	newUsers := make([]contract_helpers.UserToken, 0)
	for _, transfer := range transfers {
		token := transfer.EmitterAddress
		for _, add := range []common.Address{transfer.From, transfer.To} {
			if err, yes := utils.IsNew(ctx, block.ChainId, add, token); err == nil && yes {
				// NOTE - check if there any duplication
				utils.AddNew(ctx, block.ChainId, add, token)
				newUsers = append(newUsers, contract_helpers.UserToken{User: add, Token: token})
			} else if err != nil {
				return nil, err
			}
		}
	}
	return newUsers, nil
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

	// DEBUG
	// _bal, ok := userBal.GetBalance()
	// if !ok || _bal.Cmp(big.NewInt(0)) == -1 {
	// 	thisUserBal := schema.UserBalance{}
	// 	if res := userBalanceCol(blockTask.ChainId).FindOne(ctx, filter); res.Err() == nil {
	// 		res.Decode(&thisUserBal)
	// 	} else {
	// 		conf.Logger.Error(res.Err())
	// 	}
	// 	oldBal, _ := thisUserBal.GetBalance()
	// 	conf.Logger.Errorw("Negative Bal", "newBal", _bal.String(), "old", oldBal.String(), "block", blockTask, "user", user.String(), "tok", token.String())
	// }

	// TODO - Make this Update Many
	_, err := userBalanceCol(blockTask.ChainId).UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	return &userBal, nil
}
