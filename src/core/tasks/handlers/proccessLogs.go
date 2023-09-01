package handlers

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
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

	defer func() {
		if r := recover(); r != nil {
			conf.Logger.Infof("stacktrace from panic: %s", string(debug.Stack()))
		}
	}()

	bt := schema.BatchBlockTask{}
	if err := json.Unmarshal(task.Payload(), &bt); err != nil {
		return err
	}
	err, logs := jobs.FetchBlockEvents(ctx, bt)
	if err != nil {
		return err
	}
	helpers.SetBTFetched(ctx, bt)

	err, transfers := jobs.ProccessRawLogs(ctx, bt, logs)
	if err != nil {
		return err
	}
	helpers.SetBTParsed(ctx, bt)

	err = jobs.PrcoessParsedLogs(ctx, bt, transfers)
	if err != nil {
		return err
	}
	helpers.SetBTAdded(ctx, bt)

	// checkResults(ctx, bt)
	return nil
}

func checkResults(ctx context.Context, b schema.BatchBlockTask) {
	// NOTE: DEBUG

	col := conf.GetMongoCol(b.ChainId, conf.UserBalColName)
	eBal := contract_helpers.EasyBalanceOf{
		BlockNumber: int64(b.ToBlockNum),
		ChainId:     b.ChainId,
		UserTokens:  make([]contract_helpers.UserToken, 0),
	}

	wbnbTransfers := make([]schema.UserBalance, 0)

	curs, err := col.Find(ctx, bson.M{})
	if err != nil {
		conf.Logger.Error(err)
	} else {
		for curs.Next(ctx) {
			val := schema.UserBalance{}
			curs.Decode(&val)
			if val.TokenStr == "0xbb4CdB9CBd36B01bD1cBaEBF2De08d9173bc095c" {
				wbnbTransfers = append(wbnbTransfers, val)
				eBal.UserTokens = append(eBal.UserTokens, contract_helpers.UserToken{
					User:  val.User,
					Token: val.Token,
				})
			}
		}
	}
	conf.Logger.Infow("Fetching token's balance!")
	if len(eBal.UserTokens) > 0 {
		if err := eBal.Execute(ctx); err != nil {
			conf.Logger.Error(err)
		} else {
			for i, uBal := range eBal.UserTokens {
				our := wbnbTransfers[i]
				their := uBal
				bal, ok := our.GetBalance()
				if !ok {
					conf.Logger.Warnw("getbal failed", "o", our, "t", their)
					continue
				}
				if bal.Cmp(their.Balance) != 0 {
					conf.Logger.Warnw("Miss Match balance", "ourBal", bal, "theirBal", their.Balance, "our", our, "their", their)
				}
			}
		}
	}
}
