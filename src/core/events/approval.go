package events

import (
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (e *EventParser) ParseApprovalGroup(RawLog types.Log) (*schema.LogApproval, error) {
	return ApproveEventParser(RawLog)
}

func ApproveEventParser(vLog types.Log) (*schema.LogApproval, error) {
	var log schema.LogApproval

	err := ERC20_ABI.UnpackIntoInterface(&log, string(ApprovalE), vLog.Data)
	if err != nil {
		return nil, err
	}
	log.TokenOwner = common.HexToAddress(vLog.Topics[1].Hex())
	log.Spender = common.HexToAddress(vLog.Topics[2].Hex())
	log.TokensStr = log.Tokens.String()
	log.EmitterAddress = vLog.Address
	log.Name = string(ApprovalE)
	log.Status = schema.Fetched
	log.BlockNumber = vLog.BlockNumber
	log.TrxHash = vLog.TxHash
	log.TrxIndex = vLog.TxIndex
	log.LogIndex = vLog.Index
	return &log, nil
}
