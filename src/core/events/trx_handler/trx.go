package trx_handler

import (
	"context"
	"fmt"

	"github.com/PiperFinance/BS/src/core/schema"
)

type UserTrxHandler struct {
	Ctx       context.Context
	transfers []*schema.LogTransfer
	users     []*schema.User
}

func (h *UserTrxHandler) Flush(ctx context.Context, chainId int64, blockNumber uint64) error {
	return h.submitAllTransfers(ctx, chainId, blockNumber, h.transfers)
}

func (h *UserTrxHandler) Handle(vLog interface{}) error {
	log, ok := vLog.(*schema.LogTransfer)
	if !ok {
		return fmt.Errorf("cast error of type %T to schema.LogTransfer for %+v", vLog, vLog)
	} else if log == nil {
		return fmt.Errorf("cast error log can not be nil", vLog, vLog)
	}
	// for _, tr := range h.transfers {
	// 	if log.LogIndex == tr.LogIndex && log.BlockNumber == tr.BlockNumber {
	// 		conf.Logger.Errorw("Duplicate LogIndex", "logindex", log.LogIndex, "block", log.BlockNumber)
	// 		return nil
	// 	}
	// }
	h.transfers = append(h.transfers, log)
	return nil
}
