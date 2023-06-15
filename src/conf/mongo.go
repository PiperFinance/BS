package conf

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// LogColName Collection name for transfers events
	LogColName          = "Logs"
	BlockColName        = "Blocks"
	TokenColName        = "Tokens"
	ParsedLogColName    = "ParsedLogs"
	UserBalColName      = "UsersBalance"
	BannedUsersColName  = "BannedUsers"
	TokenVolumeColName  = "TokenVolume"
	TokenUserMapColName = "TokenUserMap"
	UserTokenMapColName = "UserTokenMap"
	QueueErrorsColName  = "QErr"
	BlockScannerDB      = "BS_Main"
	AggregatedUsers     = "Users"
)

var (
	mongoCl *mongo.Client
	// mongoDB            *mongo.Database
	MongoDefaultErrCol *mongo.Collection
	// Compound Index
	// chainIndexed map[int64]map[string]bool

)

func LoadMongo() {
	time.Sleep(Config.MongoSlowLoading)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	opts := options.Client().ApplyURI(Config.MongoUrl.String())
	opts.MaxPoolSize = &Config.MongoMaxPoolSize
	var err error
	mongoCl, err = mongo.Connect(ctx, opts)
	if err != nil {
		Logger.Panicf("Mongo: %s", err)
	}

	err = mongoCl.Ping(ctx, nil)
	if err != nil {
		Logger.Panicf("Mongo: %s", err)
	}
	// mongoDB = mongoCl.Database(Config.MongoDBName)
	MongoDefaultErrCol = mongoCl.Database(BlockScannerDB).Collection(QueueErrorsColName)
	// chainIndexed = make(map[int64]map[string]bool)
	for _, chain := range Config.SupportedChains {
		// chainIndexed[chain] = make(map[string]bool)
		GetMongoCol(chain, UserBalColName).Indexes().CreateOne(
			ctx, mongo.IndexModel{Keys: bson.D{{Key: "user", Value: 1}, {Key: "token", Value: 1}}})
		// GetMongoCol(chain, UserBalColName).Indexes().CreateOne(
		// 	ctx, mongo.IndexModel{Keys: bson.D{{Key: "user", Value: -1}, {Key: "token", Value: -1}}})
		GetMongoCol(chain, BlockColName).Indexes().CreateOne(
			ctx, mongo.IndexModel{Keys: bson.D{{Key: "no", Value: -1}}})
		GetMongoCol(chain, BlockColName).Indexes().CreateOne(
			ctx, mongo.IndexModel{Keys: bson.D{{Key: "no", Value: -1}, {Key: "status", Value: 1}}})
		GetMongoCol(chain, TokenColName).Indexes().CreateOne(
			ctx, mongo.IndexModel{Keys: bson.D{{Key: "address", Value: 1}}})
	}
}

func GetMongoCol(chain int64, colName string) *mongo.Collection {
	return mongoCl.Database(fmt.Sprintf("%s_%d", Config.MongoDBName, chain)).Collection(colName)
}
