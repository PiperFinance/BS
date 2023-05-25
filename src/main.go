package main

import (
	"context"
	"math/big"
	"time"

	"github.com/PiperFinance/BS/src/core/conf"
	"github.com/charmbracelet/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/joho/godotenv/autoload"
)

type BlockTask struct {
	BlockNumber uint64
}

func init() {
	conf.LoadConfig("./")
	conf.LoadMainNets()
	conf.LoadNetwork()
	conf.LoadQueue()
	conf.LoadMongo()
	conf.LoadRedis()
	conf.LoadDebugItems()
}

func logErr(s time.Time, rpc string, err error) {
	eStr := err.Error()
	if len(eStr) > 15 {
		eStr = eStr[:15]
	}
	log.Errorf("[%d]ms\t\t%s\t@t%s", time.Now().Sub(s).Milliseconds(), eStr, rpc)
}

func logOk(s time.Time, rpc string, block int) {
	log.Infof("[%d]ms\t\t%d\t@t%s", time.Now().Sub(s).Milliseconds(), block, rpc)
}

func test(network *conf.Network) {
	c, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	for _, rpc := range network.Rpc {
		go func(_rpc string) {
			if len(_rpc) < 1 {
				return
			}
			s := time.Now()
			if cl, err := ethclient.Dial(_rpc); err == nil {
				if block, err := cl.BlockNumber(c); err != nil {
					logErr(s, _rpc, err)
				} else {
					blockNumBigInt := big.NewInt(int64(block))
					if logs, err := cl.FilterLogs(c,
						ethereum.FilterQuery{
							FromBlock: blockNumBigInt,
							ToBlock:   blockNumBigInt,
						}); err != nil {
						logErr(s, _rpc, err)
					} else {
						logOk(s, _rpc, len(logs))
					}
				}
			} else {
				logErr(s, _rpc, err)
			}
		}(rpc.Url)
	}
	time.Sleep(15 * time.Second)
	cancel()
	log.Infof("---------------------------> %s\n", network.Name)
}

// ONLY FOR TESTING PURPOSES ...
func main() {
	// (&StartConf{}).StartAll()
	// select {}
	test(conf.ETHNetwork)
	test(conf.BSCNetwork)
	test(conf.FTMNetwork)
	test(conf.PolygonNetwork)
}
