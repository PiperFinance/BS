package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Block struct {
	FetchedAt   time.Time       `bson:"fetched_at" json:"fetched_at"`
	BlockNumber rpc.BlockNumber `bson:"block_number" json:"block_number"`
}

type LogColl struct {
	Address     common.Address     `json:"address" bson:"address" `
	Topics      []common.Hash      `bson:"topics" json:"topics" `
	Data        []byte             `bson:"data" json:"data" `
	BlockNumber uint64             `json:"blockNumber" bson:"blockNumber"`
	BlockId     primitive.ObjectID `bson:"block_id"`
	TxHash      common.Hash        `json:"transactionHash" bson:"transactionHash" `
	TxIndex     uint               `bson:"transactionIndex" json:"transactionIndex"`
	BlockHash   common.Hash        `json:"blockHash" bson:"blockHash"`
	Index       uint               `bson:"logIndex" json:"logIndex"`
	Removed     bool               `json:"removed" bson:"removed"`
}
