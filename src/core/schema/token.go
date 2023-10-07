package schema

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type TokenId string

type TokenDet struct {
	ChainId     int64          `json:"chainId"`
	Address     common.Address `json:"address"`
	Name        string         `json:"name"`
	Symbol      string         `json:"symbol"`
	Decimals    int32          `json:"decimals"`
	Tags        []string       `json:"tags"`
	CoingeckoId string         `json:"coingeckoId"`
	LifiId      string         `json:"lifiId,omitempty"`
	ListedIn    []string       `json:"listedIn"`
	LogoURI     string         `json:"logoURI"`
	Verify      bool           `json:"verify"`
	Related     []Token        `json:"token,omitempty"`
}

type Token struct {
	Detail     TokenDet  `json:"detail"`
	PriceUSD   float64   `json:"priceUSD"`
	Balance    big.Float `json:"-"`
	Value      big.Float `json:"-"`
	BalanceStr string    `json:"balance"`
	ValueStr   string    `json:"value"`
}

type (
	TokenMapping map[TokenId]Token
	// TokenAddMapping map[common.Address]Token
)

// Copy Only copies the detail bit
func (token Token) Copy() *Token {
	return &Token{
		Detail: token.Detail,
	}
}
