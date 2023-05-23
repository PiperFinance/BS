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

func userBalanceCol() *mongo.Collection {
	return conf.MongoDB.Collection(conf.UserBalColName)
}

func TokenVolumeCol() *mongo.Collection {
	return conf.MongoDB.Collection(conf.TokenVolumeColName)
}

func getBalance(ctx context.Context, blockNumber uint64, user common.Address, token common.Address) (*big.Int, error) {
	if caller, err := contracts.NewERC20Caller(token, conf.EthClient()); err != nil {
		return nil, err
	} else {
		return caller.BalanceOf(&bind.CallOpts{
			Context: ctx, BlockNumber: big.NewInt(int64(blockNumber - 1)),
		}, user)
	}
}

// UpdateUserBalTaskHandler Updates Online User's Balance and then vacumes log record from database to save space
func UpdateUserBalTaskHandler(ctx context.Context, task *asynq.Task) error {
	ctxFind, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	block := schema.BlockTask{}
	err := json.Unmarshal(task.Payload(), &block)
	if err != nil {
		log.Infof("Task ParseBlockEvents [%s] : Finished !", err)
		return err
	}
	cursor, err := conf.MongoDB.Collection(conf.ParsedLogColName).Find(ctxFind, bson.M{
		"log.blockNumber": &block.BlockNumber,
		"log.name":        events.TransferE,
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
		processTransferLog(ctx, block.BlockNumber, transfer)
		if res, err := conf.MongoDB.Collection(conf.ParsedLogColName).DeleteOne(ctxFind, bson.M{"_id": transfer.ID}); err != nil {
			log.Error(err)
		} else {
			log.Info(res)
		}
	}
	// if res, err := conf.MongoDB.Collection(conf.ParsedLogColName).DeleteMany(ctxFind, bson.M{
	// 	"log.blockNumber": &block.BlockNumber,
	// 	"log.name":        events.TransferE,
	// }); err != nil {
	// 	log.Errorf("BlockEventsTaskHandler")
	// } else {
	// 	log.Infof("Deleted Logs : %s Deleted", res.DeletedCount)
	// }
	bm := schema.BlockM{BlockNumber: block.BlockNumber}
	bm.SetParsed()
	if res, err := conf.MongoDB.Collection(conf.BlockColName).ReplaceOne(
		ctx,
		bson.M{"no": block.BlockNumber}, &bm); err != nil {
		log.Errorf("BlockEventsTaskHandler")
	} else {
		log.Infof("Replace Result : %s modified", res.ModifiedCount)
	}
	if err != nil {
		log.Errorf("Task UpdateUserBal [%d] : Err : %s !", block.BlockNumber, err)
	} else {
		log.Infof("Task UpdateUserBal [%d] : Finished !", block.BlockNumber)
	}
	return err
}

func processTransferLog(ctx context.Context, blockNumber uint64, transfer schema.LogTransfer) error {
	var amount *big.Int
	if b, ok := transfer.GetAmount(); ok {
		amount = b
	} else {
		return fmt.Errorf("transfer log get amount failure, transfer=%s", transfer)
	}
	if _, err := processUserBal(
		ctx, blockNumber,
		transfer.From, transfer.EmitterAddress,
		amount); err != nil {
		return err
	}

	if _, err := processUserBal(
		ctx, blockNumber,
		transfer.From, transfer.EmitterAddress,
		amount.Neg(amount)); err != nil {
		return err
	}

	return nil
}

func processUserBal(ctx context.Context, blockNumber uint64, user common.Address, token common.Address, amount *big.Int) (*schema.UserBalance, error) {
	userBal := schema.UserBalance{
		User:      user,
		Token:     token,
		ChangedAt: blockNumber,
	}
	filter := bson.D{{Key: "user", Value: user}, {Key: "token", Value: token}}
	if res := userBalanceCol().FindOne(ctx, filter); res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			// TODO - Get For the first time ...
			fmt.Println(user.String(), token.String())
			bal, err := getBalance(ctx, blockNumber, user, token)
			if err != nil {
				return nil, err
			}
			if conf.CallCount != nil {
				conf.CallCount.Add()
			}
			userBal.SetBalance(bal)
			userBal.StartedAt = blockNumber
			userBalanceCol().InsertOne(ctx, &userBal)
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
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "bal", Value: userBal.GetBalanceStr()}, {Key: "c_t", Value: blockNumber}}}}
	res, err := userBalanceCol().UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	log.Infof("Modified: %d , MatchedCount: %d ", res.ModifiedCount, res.MatchedCount)
	return &userBal, nil
}
