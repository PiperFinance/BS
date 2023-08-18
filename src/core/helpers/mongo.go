package helpers

import (
	"context"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/schema"
	"go.mongodb.org/mongo-driver/bson"
)

func SetBTFetched(ctx context.Context, bt schema.BatchBlockTask) {
	for i := bt.FromBlockNum; i <= bt.ToBlockNum; i++ {
		bm := schema.BlockM{BlockNumber: i, ChainId: bt.ChainId}
		bm.SetFetched()
		if _, err := conf.GetMongoCol(bt.ChainId, conf.BlockColName).ReplaceOne(
			ctx,
			bson.M{"no": i}, &bm); err != nil {
			conf.Logger.Errorf("Task FetchBlockEvents [%+v] %s", bt, err)
		}
	}
}

func SetBTParsed(ctx context.Context, bt schema.BatchBlockTask) {
	for i := bt.FromBlockNum; i <= bt.ToBlockNum; i++ {
		bm := schema.BlockM{BlockNumber: i, ChainId: bt.ChainId}
		bm.SetParsed()
		if _, err := conf.GetMongoCol(bt.ChainId, conf.BlockColName).ReplaceOne(
			ctx,
			bson.M{"no": i}, &bm); err != nil {
			conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", bt, err)
		}
	}
}

func SetBTAdded(ctx context.Context, bt schema.BatchBlockTask) {
	for i := bt.FromBlockNum; i <= bt.ToBlockNum; i++ {
		bm := schema.BlockM{BlockNumber: i, ChainId: bt.ChainId}
		bm.SetAdded()
		if _, err := conf.GetMongoCol(bt.ChainId, conf.BlockColName).ReplaceOne(
			ctx,
			bson.M{"no": i}, &bm); err != nil {
			conf.Logger.Errorf("Task ParseBlockEvents [%+v] %s", bt, err)
		}
	}
}
