package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/contracts"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

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

// UpdateUserBalTaskHandler Updates Online User's Balance and then vacuums log record from database to save space
func UpdateUserBalTaskHandler(ctx context.Context, task *asynq.Task) error {
	// TODO - Why fixed timeout ?
	ctxFind, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	blockTask := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		// log.Infof("Task ParseBlockEvents [%s] : Finished !", err)
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
	for cursor.Next(ctx) {
		transfer := schema.LogTransfer{}
		if err := cursor.Decode(&transfer); err != nil {
			log.Error(err)
			continue
		}
		processTransferLog(ctx, blockTask, transfer)
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteOne(ctxFind, bson.M{"_id": transfer.ID}); err != nil {
			log.Error(err)
		} else {
			// log.Info(res)
		}
	}
	// if res, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteMany(ctxFind, bson.M{
	// 	"log.blockNumber": &block.BlockNumber,
	// 	"log.name":        events.TransferE,
	// }); err != nil {
	// 	log.Errorf("BlockEventsTaskHandler")
	// } else {
	// 	log.Infof("Deleted Logs : %s Deleted", res.DeletedCount)
	// }
	bm := schema.BlockM{BlockNumber: blockTask.BlockNumber}
	bm.SetParsed()
	if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
		ctx,
		bson.M{"no": blockTask.BlockNumber}, &bm); err != nil {
		log.Errorf("BlockEventsTaskHandler")
	} else {
		// log.Infof("Replace Result : %s modified", res.ModifiedCount)
	}
	if err != nil {
		log.Errorf("Task UpdateUserBal [%d] : Err : %s !", blockTask.BlockNumber, err)
	} else {
		// log.Infof("Task UpdateUserBal [%d] : Finished !", block.BlockNumber)
	}
	return err
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
		if res.Err() == mongo.ErrNoDocuments {
			// TODO - Get For the first time ...
			bal, err := getBalance(ctx, blockTask, user, token)
			if err != nil {
				return nil, err
			}
			if conf.CallCount != nil {
				conf.CallCount.Add()
			}
			userBal.SetBalance(bal)
			userBal.StartedAt = blockTask.BlockNumber
			userBalanceCol(blockTask.ChainId).InsertOne(ctx, &userBal)
		} else {
			return nil, res.Err()
		}
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
	// log.Infof("Modified: %d , MatchedCount: %d ", res.ModifiedCount, res.MatchedCount)
	return &userBal, nil
}
