package handlers

import (
	"context"
	"encoding/json"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UpdateUserBalTaskHandler Updates Online User's Balance and then vacuums log record from database to save space
func UpdateUserBalTaskHandler(ctx context.Context, task *asynq.Task) error {
	// TODO - Why fixed timeout ?

	ctxInsert, cancelInsert := context.WithTimeout(ctx, 5*time.Minute)
	ctxFind, cancelFind := context.WithTimeout(ctx, 5*time.Minute)
	ctxDel, cancelDel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelDel()
	defer cancelFind()
	defer cancelInsert()
	blockTask := schema.BatchBlockTask{}
	err := json.Unmarshal(task.Payload(), &blockTask)
	if err != nil {
		return err
	}

	filter := bson.M{
		"log.blockNumber": bson.D{{Key: "$gte", Value: &blockTask.FromBlockNumber}, {Key: "$lte", Value: &blockTask.ToBlockNumber}},
		"log.name":        events.TransferE,
	}

	cursor, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).Find(ctxFind, filter)
	defer func() {
		if err := cursor.Close(ctxFind); err != nil {
			conf.Logger.Error(err)
		}
	}()
	if err != nil {
		return err
	}
	i := 0
	blockTransfers := make(map[uint64][]schema.LogTransfer)
	indicesToStore := make(map[uint64][]int, 0)
	IdsToVacuum := make([]primitive.ObjectID, 0)

	for cursor.Next(ctx) {
		transfer := schema.LogTransfer{}
		if err := cursor.Decode(&transfer); err != nil {
			conf.Logger.Errorw("UserBal", "err", err, "block", blockTask)
			continue
		}
		_, ok := blockTransfers[transfer.BlockNumber]
		if !ok {
			blockTransfers[transfer.BlockNumber] = make([]schema.LogTransfer, 0)
			indicesToStore[transfer.BlockNumber] = make([]int, 0)
		}
		IdsToVacuum = append(IdsToVacuum, transfer.ID)
		amount, ok := transfer.GetAmount()
		if ok && amount.Cmp(big.NewInt(0)) >= 1 {
			blockTransfers[transfer.BlockNumber] = append(blockTransfers[transfer.BlockNumber], transfer)
		}
		if utils.IsRegistered(transfer.From) || utils.IsRegistered(transfer.To) {
			indicesToStore[transfer.BlockNumber] = append(indicesToStore[transfer.BlockNumber], i)
		}
		i++
	}

	for blockNum, _transfers := range blockTransfers {
		if len(_transfers) > 0 {
			conf.Logger.Infof("processing [%d] block transfers", blockNum)
			if err := processTransferLogs(ctx, blockTask, _transfers); err != nil {
				return err
			}
		}

		// DEBUG - After running this shows no sign of a negative value
		for blockNum, _transfers := range blockTransfers {
			for _, transfer := range _transfers {
				amount, ok := transfer.GetAmount()
				if !ok || amount.Cmp(big.NewInt(0)) == -1 {
					conf.Logger.Errorw("BT", "no", blockNum, "tx", transfer)
				} else {
					conf.Logger.Infow("BT", "no", blockNum, "tx", transfer)
				}
			}
		}
	}

	if len(IdsToVacuum) > 0 {
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteMany(ctxDel, bson.M{"_id": bson.M{"$in": IdsToVacuum}}); err != nil {
			return err
		}
	}
	for blockNum, idx := range indicesToStore {
		if len(idx) > 0 {
			tmp := make([]interface{}, 0)
			for _, j := range idx {
				if len(blockTransfers[blockNum]) > j {
					blockTransfers[blockNum][j].ID = primitive.NilObjectID
					tmp = append(tmp, blockTransfers[blockNum][j])
				}
			}
			if _, err := conf.GetMongoCol(blockTask.ChainId, conf.TransfersColName).InsertMany(ctxInsert, tmp); err != nil {
				return err
			}
		}
	}
	for i := blockTask.FromBlockNumber; i <= blockTask.ToBlockNumber; i++ {
		bm := schema.BlockM{BlockNumber: i, ChainId: blockTask.ChainId}
		bm.SetAdded()
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.BlockColName).ReplaceOne(
			ctx,
			bson.M{"no": i}, &bm); err != nil {
			return err
		}
	}
	return err
}
