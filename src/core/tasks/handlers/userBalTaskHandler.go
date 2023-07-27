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

func UpdateUserBalTaskHandlerOrg(ctx context.Context, task *asynq.Task) error {
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
		// conf.Logger.Infof("Task ParseBlockEvents [%s] : Finished !", err)
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
	transfers := make([]schema.LogTransfer, 0)
	indicesToStore := make([]int, 0)
	IdsToVacuum := make([]primitive.ObjectID, 0)
	for cursor.Next(ctx) {
		transfer := schema.LogTransfer{}
		if err := cursor.Decode(&transfer); err != nil {
			conf.Logger.Errorw("UserBal", "err", err, "block", blockTask)
			continue
		}
		IdsToVacuum = append(IdsToVacuum, transfer.ID)
		amount, ok := transfer.GetAmount()
		if ok && amount.Cmp(big.NewInt(0)) >= 1 {
			transfers = append(transfers, transfer)
		}
		if utils.IsRegistered(transfer.From) || utils.IsRegistered(transfer.To) {
			indicesToStore = append(indicesToStore, i)
		}
		i++
	}
	if len(transfers) > 0 {
		if err := processTransferLogs(ctx, blockTask, transfers); err != nil {
			return err
		}
	}
	if len(IdsToVacuum) > 0 {
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.ParsedLogColName).DeleteMany(ctxDel, bson.M{"_id": bson.M{"$in": IdsToVacuum}}); err != nil {
			return err
		}
	}
	if len(indicesToStore) > 0 {
		tmp := make([]interface{}, 0)
		for _, j := range indicesToStore {
			if len(transfers) > j {
				transfers[j].ID = primitive.NilObjectID
				tmp = append(tmp, transfers[j])
			}
		}
		if _, err := conf.GetMongoCol(blockTask.ChainId, conf.TransfersColName).InsertMany(ctxInsert, tmp); err != nil {
			return err
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
