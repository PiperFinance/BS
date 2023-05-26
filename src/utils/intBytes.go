package utils

import (
	"encoding/json"
	"log"
	"math/big"

	"github.com/PiperFinance/BS/src/core/schema"
)

func BlockTaskGenUnsafe(chain int64) []byte {
	x := schema.BlockTask{ChainId: chain}
	r, err := json.Marshal(x)
	if err != nil {
		log.Println(" ERR : BlockTaskGenUnsafe : %+v", err)
		return nil
	} else {
		return r
	}
}

func IntToBytes(i int64) []byte {
	if i > 0 {
		return append(big.NewInt(i).Bytes(), byte(1))
	}
	return append(big.NewInt(i).Bytes(), byte(0))
}

func BytesToInt(b []byte) int64 {
	if b[len(b)-1] == 0 {
		return -big.NewInt(0).SetBytes(b[:len(b)-1]).Int64()
	}
	return big.NewInt(0).SetBytes(b[:len(b)-1]).Int64()
}
