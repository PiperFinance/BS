package main

import (
	"fmt"

	"github.com/PiperFinance/BS/src/conf"
	_ "github.com/joho/godotenv/autoload"
)

func init() {
	// NOTE - DB Sync !
	fmt.Println("BOOT : Loading Configs ... ")
	conf.LoadConfig()
	fmt.Println("BOOT : Loading Debug Tools ... ")
	conf.LoadDebugItems()
	fmt.Println("BOOT : Loading Logger ... ")
	conf.LoadLogger()
	fmt.Println("BOOT : Loading Local Cache ... ")
	conf.LoadLocalCache()
	fmt.Println("BOOT : Loading Mongo ...")
	conf.LoadMongo()
	fmt.Println("BOOT : Loading Redis ...")
	conf.LoadRedis()
	fmt.Println("BOOT : Loading Tokens ...")
	conf.LoadTokens()
	fmt.Println("BOOT : Loading Mainnets ...")
	conf.LoadMainNets()
	fmt.Println("BOOT : Loading Networks ...")
	conf.LoadNetwork()
	fmt.Println("BOOT : Loading Q ...")
	conf.LoadQueue()
}

// ONLY FOR TESTING PURPOSES ...

func main() {
	// (&StartConf{}).StartApi()
	(&StartConf{}).StartAll()
	select {}

	// results := make(map[int64]uint64)
	// for _, chain := range conf.Config.SupportedChains {
	// 	cb, _ := conf.EthClient(chain).BlockNumber(context.Background())
	// 	var i uint64
	// 	for {
	// 		i++
	// 		cl, rpc := conf.EthClientDebug(chain)
	// 		logs, err := cl.FilterLogs(
	// 			context.Background(),
	// 			ethereum.FilterQuery{
	// 				FromBlock: big.NewInt(int64(cb - i)),
	// 				ToBlock:   big.NewInt(int64(cb)),
	// 			},
	// 		)
	// 		if err != nil {
	// 			conf.Logger.Error(err)
	// 			results[chain] = i
	// 			break
	// 		}
	// 		conf.Logger.Infow("res", "rpc", rpc, "len", len(logs), "length", i, "query", []uint64{cb, cb - i})
	// 	}
	// }
	// // 1: 1 = 0x1
	// // 250: 66 = 0x42
	// // 56: 19 = 0x13
	// // 137: 15 = 0xf
	// // 42161: 313 = 0x139
	// // 9001: 87 = 0x57
	// // 58: 2698 = 0xa8a
	// // 43114: 55 = 0x37
	// // 100: 228 = 0xe4
	// // 2021: 1 = 0x1
	// // 1284: 59 = 0x3b
	// conf.Logger.Infow("res", "M", results)
}
