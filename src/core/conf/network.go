package conf

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	EthClient     *ethclient.Client
	RPCURL        string
	StartingBlock uint64
)

func init() {
	rpc, found := os.LookupEnv("RPC_URL")
	if found {
		RPCURL = rpc
	} else {
		RPCURL = "https://eth.llamarpc.com"
	}
	st, found := os.LookupEnv("STARTING_BLOCK")
	if found {
		x, err := strconv.ParseInt(st, 10, 64)
		if err != nil {
			log.Fatalf("Network: %s", err)
		}
		StartingBlock = uint64(x)
	} else {
		StartingBlock = 10000
	}
	client, err := ethclient.Dial(RPCURL)
	if err != nil {
		log.Errorf("Client Connection Error : %s  ", err)
	}
	EthClient = client
}
