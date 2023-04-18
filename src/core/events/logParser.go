package events

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/PiperFinance/BS/src/core/contracts"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

//NOTE - Inside Solidity and contracts it's called event but inside eth lib and block it's called log !

var (
	TransferEventSig     []byte
	TransferEventSigHash common.Hash
	ApprovalEventSig     []byte
	ApprovalEventSigHash common.Hash
	// TODO - more to be added
	contractAbi abi.ABI
)

func init() {
	TransferEventSig = []byte("Transfer(address,address,uint256)")
	ApprovalEventSig = []byte("Approval(address,address,uint256)")
	TransferEventSigHash = crypto.Keccak256Hash(TransferEventSig)
	ApprovalEventSigHash = crypto.Keccak256Hash(ApprovalEventSig)
	if _contractAbi, err := abi.JSON(strings.NewReader(contracts.ERC20MetaData.ABI)); err != nil {
		log.Errorf("ParseLogs: %s", err)
		contractAbi = _contractAbi
	}

}

// EventTopicHash accepts something like "ItemSet(bytes32,bytes32)" and returns with a hash like 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4
// [Source](https://goethereumbook.org/en/event-read/)
func EventTopicHash(event string) string {
	eventSignature := []byte(event)
	hash := crypto.Keccak256Hash(eventSignature)
	return hash.Hex()
}


func ParseLog(vLog types.Log) (interface{}, error) {
	switch vLog.Topics[0].Hex() {
	case TransferEventSigHash.Hex():
		return parseTransferEvent()
	case ApprovalEventSigHash.Hex():
		return parseApproveEvent()
	case _:
		return nil, EventParserNotFound
	}
}

// ParseLogs Parsers different types of log event and store them to database
func ParseLogs(ctx context.Context, mongoCol *mongo.Collection, logCursor *mongo.Cursor) {

	nextCtx, _ := context.WithTimeout(ctx, time.Second)
	for logCursor.Next(nextCtx) {
		var vLog types.Log
		errDecode := logCursor.Decode(&vLog)
		if errDecode != nil {
			log.Errorf("ParseLogs: %s", errDecode)
		}

		fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
		fmt.Printf("Log Index: %d\n", vLog.Index)

		switch vLog.Topics[0].Hex() {
		case TransferEventSigHash.Hex():
			fmt.Printf("Log Name: Transfer\n")

			var transferEvent schema.LogTransfer

			err := contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
			if err != nil {
				log.Errorf("xLogs: %s", err)
				continue
			}
			transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
			transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
			transferEvent.TokensStringValue = transferEvent.Tokens.String()
			transferEvent.EmitterAddress = vLog.Address
			_, err = mongoCol.InsertOne(context.Background(), transferEvent)
			if err != nil {
				log.Errorf("xLogs: %s", err)
				continue
			}

		case ApprovalEventSigHash.Hex():
		fmt.Printf("\n\n")
	}
}
