package events

import (
	"strings"

	"github.com/PiperFinance/BS/src/conf"
	"github.com/PiperFinance/BS/src/contracts"
	"github.com/ethereum/go-ethereum/accounts/abi"
)

var (
	ERC20_ABI   abi.ABI
	ERC721_ABI  abi.ABI
	ERC1155_ABI abi.ABI
)

func LoadParserDeps() {
	if tmp20, err := abi.JSON(strings.NewReader(contracts.ERC20MetaData.ABI)); err != nil {
		conf.Logger.Panicf("Contract Abi Loader: %s", err)
	} else {
		ERC20_ABI = tmp20
	}
	if tmp721, err721 := abi.JSON(strings.NewReader(contracts.ERC721MetaData.ABI)); err721 != nil {
		conf.Logger.Panicf("Contract Abi Loader: %s", err721)
	} else {
		ERC721_ABI = tmp721
	}
	if tmp1155, err1155 := abi.JSON(strings.NewReader(contracts.ERC1155MetaData.ABI)); err1155 != nil {
		conf.Logger.Panicf("Contract Abi Loader: %s", err1155)
	} else {
		ERC1155_ABI = tmp1155
	}
}
