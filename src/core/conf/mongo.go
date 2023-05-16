package conf

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// LogColName Collection name for transfers events
	LogColName          = "Logs"
	ParsedLogColName    = "ParsedLogs"
	UserBalColName      = "UsersBalance"
	TokenVolumeColName  = "TokenVolume"
	TokenUserMapColName = "TokenUserMap"
	UserTokenMapColName = "UserTokenMap"
)

var (
	MongoUrl    string
	MongoDBName string
	MongoCl     *mongo.Client
	MongoDB     *mongo.Database
)

func init() {
	if url, found := os.LookupEnv("MONGO_URL"); found {
		MongoUrl = url
	} else {
		MongoUrl = "mongodb://localhost:27017"
	}
	if db, found := os.LookupEnv("MONGO_DB"); found {
		MongoDBName = db
	} else {
		MongoDBName = "TEST_BS2"
	}

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	opts := options.Client().ApplyURI(MongoUrl)

	var err error
	MongoCl, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}

	err = MongoCl.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}
	MongoDB = MongoCl.Database(MongoDBName)
}
