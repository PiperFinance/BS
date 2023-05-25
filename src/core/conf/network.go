package conf

import (
	"github.com/PiperFinance/BS/src/utils"
	"github.com/charmbracelet/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	// EthClient  *ethclient.Client
	EthClientS    []*ethclient.Client
	selectorIndex int
	clientCount   int
	rpcs          []string
)

func LoadNetwork() {
	// rpcs = strings.Split(Config.RPCUrls, ",")
	rpcs = utils.GetNetworkRpcUrls(ETHNetwork.GoodRpc)
	EthClientS = make([]*ethclient.Client, len(rpcs))
	for i, v := range rpcs {
		client, err := ethclient.Dial(v)
		if err != nil {
			log.Fatalf("Client Connection %s Error : %s  ", v, err)
		}
		EthClientS[i] = client
		// EthClient = client
	}
	clientCount = len(EthClientS)
	selectorIndex = 0
}

func EthClient() *ethclient.Client {
	defer func() {
		selectorIndex++
		if selectorIndex >= clientCount {
			selectorIndex = 0
		}
	}()
	return EthClientS[selectorIndex]
}

func EthClientDebug() (*ethclient.Client, string) {
	defer func() {
		selectorIndex++
		if selectorIndex >= clientCount {
			selectorIndex = 0
		}
	}()
	return EthClientS[selectorIndex], rpcs[selectorIndex]
}
