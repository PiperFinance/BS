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
	ScannerStatus           string
)

const (
	Fetched  ScannerStatus = "Fetched"  // NOTE - Fetched Block from net
	Scanned  ScannerStatus = "Scanned"  // NOTE - Scanned From Block
	Parsed   ScannerStatus = "Parsed"   // NOTE - Parsed and is ready for further corresponding action
	Verified ScannerStatus = "Verified" // NOTE - Parsed and does not need any further action
	Checked  ScannerStatus = "Checked"  // NOTE - After parsed and ready to check for more
	Added    ScannerStatus = "Added"    // NOTE - After refreshed user's new status in db
)
