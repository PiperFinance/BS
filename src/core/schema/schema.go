package schema

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	Parsed      bool               `json:"parsed" bson:"parsed"` // Flag for parsing
}

type Log struct {
	Name           string         `bson:"name" json:"name"`
	Status         ScannerStatus  `bson:"status" json:"status"`
	EmitterAddress common.Address `json:"address" bson:"address"` // NOTE - transaction executor
	BlockNumber    uint64         `json:"blockNumber" bson:"blockNumber"`
}

type LogTransfer struct {
	Log
	From      common.Address `bson:"from" json:"from"`
	To        common.Address `bson:"to" json:"to"`
	TokensStr string         `bson:"tokens" json:"tokens"`
	Tokens    *big.Int       `bson:"-" json:"-"`
}

func (l *LogTransfer) GetAmount() (*big.Int, bool) {
	v := big.Int{}
	return v.SetString(l.TokensStr, 10)
}

// FIXME - Conflict With ERC721 Approval Event which states which NFT in collection is approved
type LogApproval struct {
	Log
	TokenOwner common.Address `bson:"owner" json:"owner"`
	Spender    common.Address `bson:"spender" json:"spender"`
	TokensStr  string         `bson:"tokens" json:"tokens"`
	Tokens     *big.Int       `bson:"-" json:"-"`
}

func (l *LogApproval) GetAmount() (*big.Int, bool) {
	v := big.Int{}
	return v.SetString(l.TokensStr, 10)
}

// LogApprovalForAll  It's a function implemented in openzeppelin vanilla contract
type LogApprovalForAll struct {
	Log
	TokenOwner common.Address `bson:"owner" json:"owner"`
	Operator   common.Address `bson:"operator" json:"operator"`
	Approved   bool           `bson:"approved" json:"approved"`
}

type LogURL struct {
	Log
	Value  string `bson:"value" json:"value"`
	NFT_ID string `bson:"nft_id" json:"nft_id"`
}
type LogTransferBatch struct {
	Log
	Operator common.Address `bson:"operator" json:"operator"`
	From     common.Address `bson:"from" json:"from"`
	To       common.Address `bson:"to" json:"to"`
	Values   []string       `bson:"values" json:"values"`
	NFT_IDs  []string       `bson:"nft_ids" json:"nft_ids"`
}

type LogTransferSingle struct {
	Log
	Operator common.Address `bson:"operator" json:"operator"`
	From     common.Address `bson:"from" json:"from"`
	To       common.Address `bson:"to" json:"to"`
	Value    string         `bson:"value" json:"value"`
	NFT_ID   string         `bson:"nft_id" json:"nft_id"`
}
