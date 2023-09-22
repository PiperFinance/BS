package events

// func ApproveEventParser(vLog types.Log) (interface{}, error) {
// 	EventName := ApprovalE
// 	var log schema.LogApproval
//
// 	err := erc20.UnpackIntoInterface(&log, EventName, vLog.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.TokenOwner = common.HexToAddress(vLog.Topics[1].Hex())
// 	log.Spender = common.HexToAddress(vLog.Topics[2].Hex())
// 	log.TokensStr = log.Tokens.String()
// 	log.EmitterAddress = vLog.Address
// 	log.Name = EventName
// 	log.Status = schema.Fetched
// 	log.BlockNumber = vLog.BlockNumber
// 	log.TrxHash = vLog.TxHash
// 	log.TrxIndex = vLog.TxIndex
// 	log.LogIndex = vLog.Index
// 	return log, nil
// }
//
// func ApproveForAllEventParser(vLog types.Log) (interface{}, error) {
// 	EventName := ApprovalForAllE
// 	var log schema.LogApprovalForAll
//
// 	err := erc721.UnpackIntoInterface(&log, EventName, vLog.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.TokenOwner = common.HexToAddress(vLog.Topics[1].Hex())
// 	log.Operator = common.HexToAddress(vLog.Topics[2].Hex())
// 	// log.TokensStringValue = log.Tokens.String()
// 	log.EmitterAddress = vLog.Address
// 	log.Status = schema.Fetched
// 	log.Name = EventName
// 	log.BlockNumber = vLog.BlockNumber
// 	log.TrxHash = vLog.TxHash
// 	log.TrxIndex = vLog.TxIndex
// 	log.LogIndex = vLog.Index
// 	return log, nil
// }
//
// func URLEventParser(vLog types.Log) (interface{}, error) {
// 	EventName := URI_E
// 	var log schema.LogURL
//
// 	err := erc1155.UnpackIntoInterface(&log, EventName, vLog.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Value = vLog.Topics[0].String()
// 	log.NFT_ID = vLog.Topics[1].Big().String()
// 	log.EmitterAddress = vLog.Address
// 	log.Name = EventName
// 	log.Status = schema.Fetched
// 	log.BlockNumber = vLog.BlockNumber
// 	log.TrxHash = vLog.TxHash
// 	log.TrxIndex = vLog.TxIndex
// 	log.LogIndex = vLog.Index
// 	return log, nil
// }
//
// func TransferBatchEventParser(vLog types.Log) (interface{}, error) {
// 	EventName := TransferBatchE
// 	var log schema.LogTransferBatch
//
// 	err := erc1155.UnpackIntoInterface(&log, EventName, vLog.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// TODO: - How does arrays work ?
//
// 	// log.TokenOwner = common.HexToAddress(vLog.Topics[1].Hex())
// 	// log.Spender = common.HexToAddress(vLog.Topics[2].Hex())
// 	// log.TokensStringValue = log.Tokens.String()
// 	log.EmitterAddress = vLog.Address
// 	log.Name = EventName
// 	log.Status = schema.Fetched
// 	log.BlockNumber = vLog.BlockNumber
// 	log.TrxHash = vLog.TxHash
// 	log.TrxIndex = vLog.TxIndex
// 	log.LogIndex = vLog.Index
// 	return log, nil
// }
//
// func TransferSingleEventParser(vLog types.Log) (interface{}, error) {
// 	EventName := TransferSingleE
// 	var log schema.LogTransferSingle
//
// 	err := erc1155.UnpackIntoInterface(&log, EventName, vLog.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	log.Operator = common.HexToAddress(vLog.Topics[0].Hex())
// 	log.From = common.HexToAddress(vLog.Topics[1].Hex())
// 	log.To = common.HexToAddress(vLog.Topics[2].Hex())
// 	log.NFT_ID = vLog.Topics[3].Big().String()
// 	log.Value = vLog.Topics[4].Big().String()
// 	log.EmitterAddress = vLog.Address
// 	log.Status = schema.Fetched
// 	log.Name = EventName
// 	log.BlockNumber = vLog.BlockNumber
// 	log.TrxHash = vLog.TxHash
// 	log.TrxIndex = vLog.TxIndex
// 	log.LogIndex = vLog.Index
// 	return log, nil
// }
//
