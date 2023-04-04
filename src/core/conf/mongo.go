package conf

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	MongoUrl    string
	MongoCl     *mongo.Client
	MongoDBName string
)

func init() {
	MongoUrl = "mongodb://localhost:27017"
	MongoDBName = "TEST_BS"

	ctx := context.TODO()
	opts := options.Client().ApplyURI("mongodb://localhost:27017")

	var err error
	MongoCl, err = mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}

	err = MongoCl.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Mongo: %s", err)
	}

}
