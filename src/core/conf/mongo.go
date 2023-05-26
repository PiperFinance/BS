package conf

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// LogColName Collection name for transfers events
	LogColName          = "Logs"
	BlockColName        = "Blocks"
	ParsedLogColName    = "ParsedLogs"
	UserBalColName      = "UsersBalance"
	TokenVolumeColName  = "TokenVolume"
	TokenUserMapColName = "TokenUserMap"
	UserTokenMapColName = "UserTokenMap"
	QueueErrorsColName  = "QErr"
	BlockScannerDB      = "BS_Main"
)

var (
	mongoCl            *mongo.Client
	mongoDB            *mongo.Database
	MongoDefaultErrCol *mongo.Collection
)

func LoadMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	opts := options.Client().ApplyURI(Config.MongoUrl.String())

	var err error
	mongoCl, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}

	err = mongoCl.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}
	mongoDB = mongoCl.Database(Config.MongoDBName)
	MongoDefaultErrCol = mongoCl.Database(BlockScannerDB).Collection(QueueErrorsColName)
}

func GetMongoCol(chain int64, colName string) *mongo.Collection {
	return mongoCl.Database(fmt.Sprintf("CID_%d", chain)).Collection(colName)
}
