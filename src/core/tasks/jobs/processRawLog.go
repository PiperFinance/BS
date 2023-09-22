package jobs

// // ProccessRawLogs [TEMP] Parses raw logs and return them
// func ProccessRawLogs(ctx context.Context, bt schema.BatchBlockTask, blocksLogs map[uint64][]types.Log) (error, map[uint64][]interface{}) {
// 	res := make(map[uint64][]interface{})
// 	for blockNo, logs := range blocksLogs {
// 		blockTrxs := make([]interface{}, 0)
// 		for _, log := range logs {
// 			parsedLog, err := events.ParseLog(log)
// 			if err != nil {
// 				switch err.(type) {
// 				case *utils.ErrEventParserNotFound:
// 					if !conf.Config.SilenceParseErrs {
// 						conf.Logger.Errorw("ParseLogs", "err", err)
// 					}
// 					continue
// 				}
// 			}
// 			if parsedLog == nil {
// 				continue
// 			} else {
// 				blockTrxs = append(blockTrxs, parsedLog)
// 			}
// 		}
// 		res[blockNo] = blockTrxs
// 	}
// 	return nil, res
// }
