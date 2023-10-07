package events

import (
	"fmt"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (e *EventParser) ParseTransferGroup(RawLog types.Log) (log *schema.LogTransfer, err error) {
	event := RawLog.Topics[0].Hex()
	switch event {
	case DepositESigHash.Hex():
		log, err = DepositEventParser(RawLog)
	case WithdrawalESigHash.Hex():
		log, err = WithdrawalEventParser(RawLog)
	case TransferESigHash.Hex():
		log, err = TransferEventParser(RawLog)
	default:
		err = &utils.ErrEventParserNotFound{Event: event, BlockNumber: RawLog.BlockNumber, TrxIndex: RawLog.TxIndex}
	}
	if err != nil {
		return nil, nil
	}
	log.Name = string(TransferE)
	log.Status = schema.Fetched
	log.BlockNumber = RawLog.BlockNumber
	log.TrxIndex = RawLog.TxIndex
	log.LogIndex = RawLog.Index
	log.TrxHash = RawLog.TxHash
	log.EmitterAddress = RawLog.Address
	log.TokensStr = log.Tokens.String()
	return log, err
}

func WithdrawalEventParser(vLog types.Log) (*schema.LogTransfer, error) {
	var log schema.LogTransfer
	if len(vLog.Topics) < 2 {
		return nil, fmt.Errorf("expected 2 topics but got %d", len(vLog.Topics))
	}

	err := ERC20_ABI.UnpackIntoInterface(&log, string(WithdrawalE), vLog.Data)
	if err != nil {
		return nil, err
	}
	// NOTE :It's Null Address so yeah
	// log.To = common.HexToAddress(vLog.Topics[1].Hex())
	log.From = common.HexToAddress(vLog.Topics[1].Hex())
	return &log, err
}

func DepositEventParser(vLog types.Log) (*schema.LogTransfer, error) {
	var log schema.LogTransfer
	if len(vLog.Topics) < 2 {
		return nil, fmt.Errorf("expected 2 topics but got %d", len(vLog.Topics))
	}

	err := ERC20_ABI.UnpackIntoInterface(&log, string(DepositE), vLog.Data)
	if err != nil {
		return nil, err
	}
	// NOTE :It's Null Address so yeah
	// log.From = common.HexToAddress(vLog.Topics[1].Hex())
	log.To = common.HexToAddress(vLog.Topics[1].Hex())
	return &log, err
}

func TransferEventParser(vLog types.Log) (*schema.LogTransfer, error) {
	var log schema.LogTransfer
	if len(vLog.Topics) < 3 {
		return nil, fmt.Errorf("expected 3 topics but got %d", len(vLog.Topics))
	}

	err := ERC20_ABI.UnpackIntoInterface(&log, string(TransferE), vLog.Data)
	if err != nil {
		return nil, err
	}
	log.From = common.HexToAddress(vLog.Topics[1].Hex())
	log.To = common.HexToAddress(vLog.Topics[2].Hex())
	return &log, err
}
