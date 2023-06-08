package events

import (
	"context"
	"time"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"go.mongodb.org/mongo-driver/mongo"
)

// NOTE - Inside Solidity and contracts it's called event but inside eth lib and block it's called log !
const (
	// TODO - more to be added
	// Event Names
	// ERC20
	TransferE = "Transfer"
	ApprovalE = "Approval"
	// ERC721
	ApprovalForAllE = "ApprovalForAll"
	// ERC1155
	URI_E           = "URI"
	TransferBatchE  = "TransferBatch"
	TransferSingleE = "TransferSingle"
	// Event Signatures
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
	if len(vLog.Topics) > 1 {
		event := vLog.Topics[0].Hex()
		switch event {
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
			return nil, &utils.ErrEventParserNotFound{Event: event, BlockNumber: vLog.BlockNumber, TrxIndex: vLog.TxIndex}
		}
	}
	// TODO - No data
	return nil, nil
}

// ParseLogs Parsers different types of log event and store them to database
func ParseLogs(ctx context.Context, mongoCol *mongo.Collection, logCursor *mongo.Cursor) {
	nextCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	parsedLogs := make([]interface{}, 0)
	for logCursor.Next(nextCtx) {
		var vLog types.Log
		errDecode := logCursor.Decode(&vLog)
		if errDecode != nil {
			conf.Logger.Errorf("ParseLogs: [%T] :%s", errDecode, errDecode)
			continue
		}
		if parsedLog, parseErr := ParseLog(vLog); parseErr != nil {
			switch parseErr.(type) {
			case *utils.ErrEventParserNotFound:
				if !conf.Config.SilenceParseErrs {
					conf.Logger.Errorf("ParseLogs: [%T] : %s", parseErr, parseErr)
				}
				continue
			}
		} else {
			if parsedLog != nil {
				parsedLogs = append(parsedLogs, parsedLog)
			}
		}
	}
	if len(parsedLogs) > 0 {
		_, insertionErr := mongoCol.InsertMany(ctx, parsedLogs)
		if insertionErr != nil {
			conf.Logger.Errorf("ParseLogs: [%T] : %s", insertionErr, insertionErr)
		}
	}
}
