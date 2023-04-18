package events

import (
	"context"
	"errors"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

//NOTE - Inside Solidity and contracts it's called event but inside eth lib and block it's called log !

const (
	// TODO - more to be added
	//Event Names
	//ERC20
	TransferE = "Transfer"
	ApprovalE = "Approval"
	//ERC721
	ApprovalForAllE = "ApprovalForAll"
	//ERC1155
	URI_E           = "URI"
	TransferBatchE  = "TransferBatch"
	TransferSingleE = "TransferSingle"
	//Event Signatures
	TransferESig       = "Transfer(address,address,uint256)"
	ApprovalESig       = "Approval(address,address,uint256)"
	ApprovalForAllESig = "ApprovalForAll(address,address,bool)"
	URI_ESig           = "URI(string,uint256)"
	TransferBatchESig  = "TransferBatch(address,address,address,uint256[],uint256[])"
	TransferSingleESig = "TransferSingle(address,address,address,uint256,uint256)"
)

var (
	transferESigHash       common.Hash
	approvalESigHash       common.Hash
	approvalForAllESigHash common.Hash
	uRIESigHash            common.Hash
	transferBatchESigHash  common.Hash
	transferSingleESigHash common.Hash

	ErrEventParserNotFound = errors.New("EventParserNotFound")
)

func init() {
	uRIESigHash = crypto.Keccak256Hash([]byte(URI_ESig))
	transferESigHash = crypto.Keccak256Hash([]byte(TransferESig))
	approvalESigHash = crypto.Keccak256Hash([]byte(ApprovalESig))
	transferBatchESigHash = crypto.Keccak256Hash([]byte(TransferBatchESig))
	approvalForAllESigHash = crypto.Keccak256Hash([]byte(ApprovalForAllESig))
	transferSingleESigHash = crypto.Keccak256Hash([]byte(TransferSingleESig))

}

// ParseLog Select Appropriate EventParser For found event !
func ParseLog(vLog types.Log) (interface{}, error) {
	switch vLog.Topics[0].Hex() {
	case transferESigHash.Hex():
		return TransferEventParser(vLog)
	case approvalESigHash.Hex():
		return ApproveEventParser(vLog)
	case approvalForAllESigHash.Hex():
		return ApproveForAllEventParser(vLog)
	case uRIESigHash.Hex():
		return URLEventParser(vLog)
	case transferBatchESigHash.Hex():
		return TransferBatchEventParser(vLog)
	case transferSingleESigHash.Hex():
		return TransferSingleEventParser(vLog)
	default:
		return nil, ErrEventParserNotFound
	}
}

// ParseLogs Parsers different types of log event and store them to database
func ParseLogs(ctx context.Context, mongoCol *mongo.Collection, logCursor *mongo.Cursor) {

	nextCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	for logCursor.Next(nextCtx) {
		var vLog types.Log
		errDecode := logCursor.Decode(&vLog)
		if errDecode != nil {
			log.Errorf("ParseLogs: [%T] :%s", errDecode, errDecode)
		}

		if parsedLog, parseErr := ParseLog(vLog); parseErr != nil {
			log.Errorf("ParseLogs: [%T] : %s", parseErr, parseErr)
		} else {
			_, insertionErr := mongoCol.InsertOne(ctx, parsedLog)
			if insertionErr != nil {
				log.Errorf("ParseLogs: [%T] : %s", insertionErr, insertionErr)
			}
		}

		log.Infof("Log Block Number: %d\n", vLog.BlockNumber)
		log.Infof("Log Index: %d\n", vLog.Index)
	}
}
