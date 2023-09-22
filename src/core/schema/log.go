package schema

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ILog interface {
	GetType() string
}

type Log struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"_id,"`
	Name           string             `bson:"name" json:"name"`
	Status         string             `bson:"status" json:"status"`
	EmitterAddress common.Address     `json:"address" bson:"address"` // NOTE: Token / Contract which emitted event
	BlockNumber    uint64             `json:"blockNumber" bson:"blockNumber"`
	TrxHash        common.Hash        `json:"txHash" bson:"txHash"`
	TrxIndex       uint               `json:"txIdx" bson:"txIdx"`
	LogIndex       uint               `bson:"logIndex" json:"logIndex"` // NOTE: index of log/event in block
}

func (l Log) GetType() string {
	return "Log"
}

type LogTransfer struct {
	Log
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	From      common.Address     `bson:"from" json:"from"`
	To        common.Address     `bson:"to" json:"to"`
	TokensStr string             `bson:"tokens" json:"tokens"`
	Tokens    *big.Int           `bson:"-" json:"-"`
}

func (l LogTransfer) GetType() string {
	return "Transfer"
}

func (l *LogTransfer) GetAmount() (*big.Int, bool) {
	v := big.Int{}
	return v.SetString(l.TokensStr, 10)
}

// FIXME: Conflict With ERC721 Approval Event which states which NFT in collection is approved
type LogApproval struct {
	Log
	TokenOwner common.Address `bson:"owner" json:"owner"`
	Spender    common.Address `bson:"spender" json:"spender"`
	TokensStr  string         `bson:"tokens" json:"tokens"`
	Tokens     *big.Int       `bson:"-" json:"-"`
}

func (l LogApproval) GetType() string {
	return "Approve"
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
