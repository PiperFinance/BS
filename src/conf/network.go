package conf

import (
	"context"
	"sync"

	"github.com/PiperFinance/BS/src/utils"
	"github.com/charmbracelet/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// EthClient  *ethclient.Client
	EthClientS    map[int64][]*ethclient.Client
	selectorMutex sync.Mutex
	selectorIndex map[int64]int
	clientCount   map[int64]int
	rpcs          map[int64][]string
)

func LoadNetwork() {
	selectorMutex = sync.Mutex{}
	rpcs = make(map[int64][]string, len(Config.SupportedChains))
	EthClientS = make(map[int64][]*ethclient.Client, len(Config.SupportedChains))
	clientCount = make(map[int64]int, len(Config.SupportedChains))
	selectorIndex = make(map[int64]int, len(Config.SupportedChains))
	for _, net := range SupportedNetworks {

		rpcs[net.ChainId] = utils.GetNetworkRpcUrls(net.GoodRpc)
		clientCount[net.ChainId] = len(net.GoodRpc)
		selectorIndex[net.ChainId] = 0
		EthClientS[net.ChainId] = make([]*ethclient.Client, len(net.GoodRpc))
		for i, _rpc := range net.GoodRpc {
			client, err := ethclient.Dial(_rpc.Url)
			if err != nil {
				log.Fatalf("Client Connection %+v Error : %s  ", _rpc, err)
			}
			EthClientS[net.ChainId][i] = client
		}
	}
}

func EthClient(chain int64) *ethclient.Client {
	defer func() {
		selectorMutex.Lock()
		selectorIndex[chain]++
		if selectorIndex[chain] >= clientCount[chain] {
			selectorIndex[chain] = 0
		}
		selectorMutex.Unlock()
	}()
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println("Recovered in f", r)
	// 	}
	// }()
	// TODO - Try to recover panic here !
	if clients, ok := EthClientS[chain]; ok {
		if index, ok := selectorIndex[chain]; ok {
			if len(clients) == 0 {
				Logger.Errorf("EthCLient Selector : No RPCs found for chain %d", chain)
			} else {
				return clients[index]
			}
		}
	}
	return nil
}

func StartingBlock(ctx context.Context, chain int64) uint64 {
	if b, err := EthClient(chain).BlockNumber(ctx); err != nil {
		return Config.StartingBlockNumber
	} else {
		return b
	}
}

// func EthClientDebug() (*ethclient.Client, string) {
// 	defer func() {
// 		selectorIndex++
// 		if selectorIndex >= clientCount {
// 			selectorIndex = 0
// 		}
// 	}()
// 	return EthClientS[selectorIndex], rpcs[selectorIndex]
// }
