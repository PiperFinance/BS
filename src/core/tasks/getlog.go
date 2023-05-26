package tasks

// func GetTransferLogsTask(ctx context.Context, t *asynq.Task) error {
// 	var blockNum uint64
// 	err := json.Unmarshal(t.Payload(), &blockNum)
// 	if err != nil {
// 		conf.Logger.Errorf("GetTransferLogsTask: %s", err)
// 	}
// 	var tokenAdd common.Address
// 	token, err := contracts.NewERC20(tokenAdd, conf.EthClient())
// 	_ = token
// 	return err
// }
