package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/PiperFinance/BS/src/core/schema"
	"github.com/ethereum/go-ethereum/common"
)

var (
	// CD chain Tokens URL
	allTokensArray = make([]schema.Token, 0)
	allTokens      = make(schema.TokenMapping)
	chainTokens    = make(map[int64]schema.TokenMapping)
	chainAddTokens = make(map[int64]map[common.Address]schema.TokenId)
)

func LoadTokens() {
	// Load Tokens ...
	var byteValue []byte
	if _, err := os.Stat(Config.TokensDir); errors.Is(err, os.ErrNotExist) {
		resp, err := http.Get(Config.TokenListUrl.String())
		if err != nil {
			Logger.Panicln(err)
		}
		byteValue, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			Logger.Panicf("HTTPTokenLoader: %s", err)
		}
	} else {
		jsonFile, err := os.Open(Config.TokensDir)
		defer func(jsonFile *os.File) {
			err := jsonFile.Close()
			if err != nil {
				Logger.Error(err)
			}
		}(jsonFile)
		if err != nil {
			Logger.Panicf("JSONTokenLoader: %s", err)
		}
		byteValue, err = ioutil.ReadAll(jsonFile)
		if err != nil {
			Logger.Panicf("JSONTokenLoader: %s", err)
		}
	}
	err := json.Unmarshal(byteValue, &allTokens)
	if err != nil {
		Logger.Panicf("TokenLoader: %s", err)
	}
	for tokenId, token := range allTokens {
		chainId := token.Detail.ChainId
		if chainTokens[chainId] == nil {
			chainTokens[chainId] = make(schema.TokenMapping)
			chainAddTokens[chainId] = make(map[common.Address]schema.TokenId)
		}
		chainTokens[chainId][tokenId] = token
		chainAddTokens[chainId][token.Detail.Address] = tokenId
		allTokensArray = append(allTokensArray, token)
	}
}

func AllChainsTokens() schema.TokenMapping {
	return allTokens
}

func AllChainsTokensArray() []schema.Token {
	return allTokensArray
}

func FindTokenId(chainId int64, add common.Address) schema.TokenId {
	id, ok := chainAddTokens[chainId][add]
	// crc32("-".join([self.address.lower(), str(self.chainId)]).encode()) <- token's Id
	key := fmt.Sprintf("%s-%d", strings.ToLower(add.String()), chainId)
	if !ok {
		sum := crc32.ChecksumIEEE([]byte(key))
		return schema.TokenId(sum)
	} else {
		return id
	}
}

func ChainTokens(id int64) schema.TokenMapping {
	t := chainTokens[id]
	return t
}
