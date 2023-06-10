package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
)

var (
	MainNets          []*schema.Network
	SupportedNetworks map[int64]*schema.Network
)

func LoadMainNets() {
	MainNets = make([]*schema.Network, 0)
	SupportedNetworks = make(map[int64]*schema.Network, len(Config.SupportedChains))
	jsonFile, err := os.Open("data/mainnets.json")
	defer jsonFile.Close()
	if err != nil {
		Logger.Fatal(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, &MainNets); err != nil {
		Logger.Fatal(err)
	}
	for _, _net := range MainNets {
		if utils.Contains(Config.SupportedChains, _net.ChainId) {
			go utils.NetworkConnectionCheck(CallCount, FailedCallCount, Logger, _net, Config.TestTimeout)
			SupportedNetworks[_net.ChainId] = _net
		}
	}
	time.Sleep(Config.TestTimeout)
	for _, chain := range Config.SupportedChains {
		sn, ok := SupportedNetworks[chain]
		if ok && len(sn.GoodRpc) < 1 {
			Logger.Fatalf("No Good Rpc for chain %d", chain)
		} else if !ok {
			Logger.Fatalf("Where is Rpc for chain %d", chain)
		}
	}
}
