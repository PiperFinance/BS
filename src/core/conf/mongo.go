package conf

import (
	"context"
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
)

var (
	MongoCl *mongo.Client
	MongoDB *mongo.Database
)

func LoadMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	opts := options.Client().ApplyURI(Config.MongoUrl)

	var err error
	MongoCl, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}

	err = MongoCl.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}
	MongoDB = MongoCl.Database(Config.MongoDBName)
}
