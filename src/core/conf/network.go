package conf

import (
	"github.com/charmbracelet/log"

	"github.com/ethereum/go-ethereum/ethclient"
)

var EthClient *ethclient.Client

func LoadNetwork() {
	client, err := ethclient.Dial(Config.RPCUrl)
	if err != nil {
		log.Fatalf("Client Connection Error : %s  ", err)
	}
	EthClient = client
}
