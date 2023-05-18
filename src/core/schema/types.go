package schema

import "net/url"

type (
	Name     string
	Symbol   string
	Decimals int32
	Price    float64

	// type Balance big.Float
	ChainId                 int64
	NetworkId               uint64
	ChainName               string
	NetworkExplorerStandard string
	RPCUrl                  url.URL
	ScannerStatus           int32
)

const (
	Fetched  ScannerStatus = iota // NOTE - Fetched Block from net
	Scanned                       // NOTE - Scanned From Block
	Parsed                        // NOTE - Parsed and is ready for further corresponding action
	Verified                      // NOTE - Parsed and does not need any further action
	Checked                       // NOTE - After parsed and ready to check for more
	Added                         // NOTE - After refreshed user's new status in db
)
