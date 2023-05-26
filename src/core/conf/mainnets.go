package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/PiperFinance/BS/src/utils"
	"github.com/charmbracelet/log"
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
		log.Fatal(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, &MainNets); err != nil {
		log.Fatal(err)
	}
	for _, _net := range MainNets {
		if utils.Contains(Config.SupportedChains, _net.ChainId) {
			go utils.NetworkConnectionCheck(_net, Config.TestTimeout)
			SupportedNetworks[_net.ChainId] = _net
		}
	}
	time.Sleep(Config.TestTimeout)
}
