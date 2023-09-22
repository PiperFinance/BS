package events

import (
	"context"

	"github.com/PiperFinance/BS/src/core/events/trx_handler"
	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/core/types"
)

type EventParser struct {
	handlers map[EventGroup][]Handler
}

func NewEventParser() *EventParser {
	ep := &EventParser{
		handlers: map[EventGroup][]Handler{
			TransferGrp: {&trx_handler.UserTrxHandler{}},
			ApprovalGrp: {},
		},
	}
	return ep
}

func (e *EventParser) AddHandler(et EventGroup, handler Handler) {
	_, ok := e.handlers[et]
	if !ok {
		e.handlers[et] = []Handler{handler}
	} else {
		e.handlers[et] = append(e.handlers[et], handler)
	}
}

func (e *EventParser) getGroup(RawLog types.Log) EventGroup {
	if len(RawLog.Topics) < 1 {
		return ""
	}
	event := RawLog.Topics[0].Hex()
	switch event {
	case TransferESigHash.Hex(), DepositESigHash.Hex(), WithdrawalESigHash.Hex():
		return TransferGrp
	case ApprovalESigHash.Hex():
		return ApprovalGrp
	// 	return ApproveEventParser(vLog)
	// case approvalForAllESigHash.Hex():
	// 	return ApproveForAllEventParser(vLog)
	// case uRIESigHash.Hex():
	// 	return URLEventParser(vLog)
	// case transferBatchESigHash.Hex():
	// 	return TransferBatchEventParser(vLog)
	// case transferSingleESigHash.Hex():
	// 	return TransferSingleEventParser(vLog)
	default:
		return ""
	}
	// NOTE: Maybe you want to log this
	// &utils.ErrEventParserNotFound{Event: RawLog, BlockNumber: vLog.BlockNumber, TrxIndex: vLog.TxIndex}
}

func (e *EventParser) handleAprvGrp(log *schema.LogApproval) error {
	for _, handler := range e.handlers[ApprovalGrp] {
		if err := handler.Handle(log); err != nil {
			return err
		}
	}
	return nil
}

func (e *EventParser) handleTrGrp(log *schema.LogTransfer) error {
	for _, handler := range e.handlers[TransferGrp] {
		if log == nil {
			continue
		}
		if err := handler.Handle(log); err != nil {
			return err
		}
	}
	return nil
}

func (e *EventParser) Flush(ctx context.Context, chainId int64, blockNumber uint64) error {
	for _, handlers := range e.handlers {
		for _, handler := range handlers {
			if err := handler.Flush(ctx, chainId, blockNumber); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *EventParser) Parse(RawLog types.Log) error {
	if len(RawLog.Data) == 0 {
		return nil
	}
	switch e.getGroup(RawLog) {
	case "":
		return nil
	case ApprovalGrp:
		l, err := e.ParseApprovalGroup(RawLog)
		if err != nil {
			return err
		}
		return e.handleAprvGrp(l)
	case TransferGrp:
		l, err := e.ParseTransferGroup(RawLog)
		if err != nil {
			return err
		}
		return e.handleTrGrp(l)
	}
	return nil
}
