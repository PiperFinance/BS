package jobs

// // BPPL : Block Process Parse Log
// type BPPL struct {
// 	bt        schema.BlockTask
// 	transfers []schema.LogTransfer
// 	// TODO: Add Other types here
// }
//
// func (l *BPPL) submit(c context.Context) error {
// 	return submitAllTransfers(c, l.bt, l.transfers)
// }
//
// // PrcoessParsedLogs takes parsed log interfaces and do actions needed for each on based on their types
// func PrcoessParsedLogs(ctx context.Context, bt schema.BatchBlockTask, blockParsedLogs map[uint64][]interface{}) error {
// 	for blockNum := bt.FromBlockNum; blockNum <= bt.ToBlockNum; blockNum++ {
// 		bppl := BPPL{bt: schema.BlockTask{BlockNumber: blockNum, ChainId: bt.ChainId}}
// 		logs, ok := blockParsedLogs[blockNum]
// 		if !ok {
// 			conf.Logger.Warnw("ProccessParsedLogs: missing key in parsed map", "blockNum", blockNum)
// 			continue
// 		}
// 		for _, log := range logs {
// 			tr, ok := log.(schema.LogTransfer)
// 			if ok {
//
// 				// NOTE: DEBUG Binance Hot Wallet
// 				BHW := "0x3c783c21a0383057D128bae431894a5C19F9Cf06"
// 				if tr.From.String() == BHW || tr.To.String() == BHW {
// 					if _, err := conf.GetMongoCol(bt.ChainId, conf.TransfersColName).InsertOne(ctx, tr); err != nil {
// 						conf.Logger.Errorw("TransferInsertion", "err", err, "tr", tr)
// 						// return err
// 					}
// 				}
// 				bppl.transfers = append(bppl.transfers, tr)
// 			}
// 		}
// 		bppl.submit(ctx)
// 	}
// 	return nil
// }
