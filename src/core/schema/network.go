package schema

type RPC struct {
	Url             string `json:"url"`
	Tracking        string `json:"tracking,omitempty"`
	TrackingDetails string `json:"trackingDetails,omitempty"`
	IsOpenSource    bool   `json:"isOpenSource,omitempty"`
}

type Network struct {
	Name     string `json:"name"`
	Chain    string `json:"chain"`
	Icon     string `json:"icon,omitempty"`
	Rpc      []RPC  `json:"rpc"`
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
	// NOTE - is calculated after running connection check
	GoodRpc           []*RPC
	BadRpc            []*RPC
	BatchLogMaxHeight int64 `json:"maxGetLogHeight"`  // GetLogs Filter max length can be updated but initial value is set in the Mainnets.json
	MulticallMaxSize  int64 `json:"maxMulticallSize"` // It's kinda obvious :)
}
