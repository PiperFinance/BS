package tasks

import (
	"context"
	"encoding/json"
	"github.com/PiperFinance/BS/src/core/conf"
	ERC20 "github.com/PiperFinance/BS/src/core/contracts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
	log "github.com/sirupsen/logrus"
)

func GetTransferLogsTask(ctx context.Context, t *asynq.Task) error {
	var blockNum uint64
	err := json.Unmarshal(t.Payload(), &blockNum)
	if err != nil {
		log.Errorf("GetTransferLogsTask: %s", err)
	}
	var tokenAdd common.Address
	token, err := ERC20.NewERC20(tokenAdd, conf.EthClient)
	//conf.EthClient.
	//token.ParseTransfer(nil)
	//token.tra
	_ = token
	return err
}
