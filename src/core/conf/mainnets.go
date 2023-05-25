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
	MainNets       []*schema.Network
	ETHNetwork     *schema.Network
	PolygonNetwork *schema.Network
	FTMNetwork     *schema.Network
	BSCNetwork     *schema.Network
)

func LoadMainNets() {
	MainNets = make([]*schema.Network, 0)
	jsonFile, err := os.Open("data/mainnets.json")
	defer jsonFile.Close()
	if err != nil {
		log.Fatal(err)
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)
	if err := json.Unmarshal(byteValue, &MainNets); err != nil {
		log.Fatal(err)
	}
	TestTimeout := 30 * time.Second
	for _, _net := range MainNets {
		switch _net.ChainId {
		case 1:
			ETHNetwork = _net
			go utils.NetworkConnectionCheck(ETHNetwork, TestTimeout)
		case 56:
			BSCNetwork = _net
			go utils.NetworkConnectionCheck(BSCNetwork, TestTimeout)
		case 137:
			PolygonNetwork = _net
			go utils.NetworkConnectionCheck(PolygonNetwork, TestTimeout)
		case 250:
			FTMNetwork = _net
			go utils.NetworkConnectionCheck(FTMNetwork, TestTimeout)
		}
	}
	time.Sleep(TestTimeout)
}
