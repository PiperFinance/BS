package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/core/helpers"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/core/tasks/jobs"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

// ProccessBlockTaskHandler Uses FetchBlockEventsKey and requires BlockTask as arg
// Calls for events and store them to mongo !
func ProccessBlockTaskHandler(ctx context.Context, task *asynq.Task) error {
	// TODO: Add check for task retries X_X
	bt := schema.BatchBlockTask{}
	if err := json.Unmarshal(task.Payload(), &bt); err != nil {
		return err
	}
	err, logs := jobs.FetchBlockEvents(ctx, bt)
	if err != nil {
		return err
	}
	err, transfers := jobs.ProccessRawLogs(ctx, bt, logs)
	if err != nil {
		return err
	}
	err = jobs.PrcoessParsedLogs(ctx, bt, transfers)
	helpers.SetBTFetched(ctx, bt)
	checkResults(ctx, bt)
	return nil
}

func checkResults(ctx context.Context, b schema.BatchBlockTask) {
	// NOTE: DEBUG
	col := conf.GetMongoCol(b.ChainId, conf.UserBalColName)

	curs, err := col.Find(ctx, bson.M{})
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
					fmt.Println("NEGATIVE BAL")
				}
			}
		}
	}
}
