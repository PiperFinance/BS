package conf

import (
	log "github.com/sirupsen/logrus"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	EthClient *ethclient.Client
	RPCURL    string
)

func init() {
	RPCURL = "https://eth.llamarpc.com"
	client, err := ethclient.Dial(RPCURL)
	if err != nil {
		log.Errorf("Client Connection Error : %s  ", err)
	}
	EthClient = client
}
