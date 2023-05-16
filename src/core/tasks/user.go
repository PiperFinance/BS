package tasks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/events"
	"github.com/charmbracelet/log"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
)

// UpdateUserBalTaskHandler Updates Online User's Balance and then vacumes log record from database to save space
func UpdateUserBalTaskHandler(ctx context.Context, task *asynq.Task) error {
	log.Infof("Task blockScan : Started !")

	block := BlockTask{}
	mongoParsedLogsCol := conf.MongoDB.Collection(conf.ParsedLogColName)
	err := json.Unmarshal(task.Payload(), &block)
	if err != nil {
		log.Infof("Task ParseBlockEvents [%s] : Finished !", err)
		return err
	}
	ctxFind, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cursor, err := conf.MongoDB.Collection(conf.LogColName).Find(ctxFind, bson.M{"blockNumber": &block.BlockNumber})
	defer cursor.Close(ctxFind)
	if err != nil {
		return err
	}
	events.ParseLogs(ctx, mongoParsedLogsCol, cursor)
	log.Infof("Task ParseBlockEvents [%s] : Finished !", err)

	return err

	log.Infof("Task blockScan [%s] : Finished !", err)
	return err
}
