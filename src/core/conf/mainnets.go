package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/log"
)

var (
	MainNets       []*Network
	ETHNetwork     *Network
	PolygonNetwork *Network
	FTMNetwork     *Network
	BSCNetwork     *Network
)

type Network struct {
	Name  string `json:"name"`
	Chain string `json:"chain"`
	Icon  string `json:"icon,omitempty"`
	Rpc   []struct {
		Url             string `json:"url"`
		Tracking        string `json:"tracking,omitempty"`
		TrackingDetails string `json:"trackingDetails,omitempty"`
		IsOpenSource    bool   `json:"isOpenSource,omitempty"`
	} `json:"rpc"`
	Features []struct {
		Name string `json:"name"`
	} `json:"features,omitempty"`
	Faucets        []string `json:"faucets"`
	NativeCurrency struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals int    `json:"decimals"`
	} `json:"nativeCurrency"`
	InfoURL   string `json:"infoURL"`
	ShortName string `json:"shortName"`
	ChainId   int64  `json:"chainId"`
	NetworkId int64  `json:"networkId"`
	Slip44    int64  `json:"slip44,omitempty"`
	Ens       struct {
		Registry string `json:"registry"`
	} `json:"ens,omitempty"`
	Explorers []struct {
		Name     string `json:"name"`
		Url      string `json:"url"`
		Standard string `json:"standard"`
		Icon     string `json:"icon,omitempty"`
	} `json:"explorers,omitempty"`
	Tvl       float64 `json:"tvl,omitempty"`
	ChainSlug string  `json:"chainSlug,omitempty"`
	Parent    struct {
		Type    string `json:"type"`
		Chain   string `json:"chain"`
		Bridges []struct {
			Url string `json:"url"`
		} `json:"bridges,omitempty"`
	} `json:"parent,omitempty"`
	Title    string   `json:"title,omitempty"`
	Status   string   `json:"status,omitempty"`
	RedFlags []string `json:"redFlags,omitempty"`
}

func LoadMainNets() {
	MainNets = make([]*Network, 0)
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
		switch _net.ChainId {
		case 1:
			ETHNetwork = _net
		case 56:
			BSCNetwork = _net
		case 137:
			PolygonNetwork = _net
		case 250:
			FTMNetwork = _net
			// case 1:
			// 	ETHNetwork = &_net
			// case 1:
			// 	ETHNetwork = &_net

		}
	}
}
