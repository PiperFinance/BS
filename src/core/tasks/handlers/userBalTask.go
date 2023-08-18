package handlers

import (
	"context"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func queryLogsForTransfers(ctx context.Context, bt schema.BatchBlockTask) (
	blockTransfers map[uint64][]schema.LogTransfer,
	indicesToStore map[uint64][]int,
	IdsToVacuum []primitive.ObjectID,
	err error,
) {
	filter := bson.M{
		"log.blockNumber": bson.D{{Key: "$gte", Value: &bt.FromBlockNum}, {Key: "$lte", Value: &bt.ToBlockNum}},
		"log.name":        events.TransferE,
	}

	ctxFind, cancelFind := context.WithTimeout(ctx, 5*time.Minute)
	defer cancelFind()
	cursor, err := conf.GetMongoCol(bt.ChainId, conf.ParsedLogColName).Find(ctxFind, filter)
	defer func() {
		if err := cursor.Close(ctxFind); err != nil {
			conf.Logger.Error(err)
		}
	}()
	if err != nil {
		return
	}

	blockTransfers = make(map[uint64][]schema.LogTransfer)
	indicesToStore = make(map[uint64][]int, 0)
	IdsToVacuum = make([]primitive.ObjectID, 0)

	for cursor.Next(ctx) {
		transfer := schema.LogTransfer{}
		if err := cursor.Decode(&transfer); err != nil {
			conf.Logger.Errorw("UserBal", "err", err, "block", bt)
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
			indicesToStore[transfer.BlockNumber] = append(indicesToStore[transfer.BlockNumber], len(blockTransfers[transfer.BlockNumber])-1)
		}
	}
	return
}

func updateUserBalJob(ctx context.Context, bt schema.BatchBlockTask, blockTransfers map[uint64][]schema.LogTransfer) error {
	for _, blockNum := range utils.SortedKeys[uint64, []schema.LogTransfer](blockTransfers) {
		_transfers, ok := blockTransfers[blockNum]
		if !ok {
			continue
		}
		if len(_transfers) > 0 {
			conf.Logger.Infof("processing [%d] block transfers", blockNum)
			thisBlock := schema.BlockTask{
				BlockNumber: blockNum,
				ChainId:     bt.ChainId,
			}
			if err := processTransferLogs(ctx, thisBlock, _transfers); err != nil {
				return err
			}
		}
	}

	return nil
}
