package tasks

import (
	"context"
	"encoding/json"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/PiperFinance/BS/src/core/contracts"
	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
)

func GetTransferLogsTask(ctx context.Context, t *asynq.Task) error {
	var blockNum uint64
	err := json.Unmarshal(t.Payload(), &blockNum)
	if err != nil {
		log.Errorf("GetTransferLogsTask: %s", err)
	}
	var tokenAdd common.Address
	token, err := contracts.NewERC20(tokenAdd, conf.EthClient)
	_ = token
	return err
}
