package utils

import (
	"github.com/PiperFinance/BS/src/conf"
	contract_helpers "github.com/PiperFinance/BS/src/contracts/helpers"
)

func ChunkArr[T any](users []T, size uint64) [][]T {
	batchSize := int(size)
	chunkCount := (len(users) / batchSize) + 1
	r := make([][]T, chunkCount)
	for i := 0; i < chunkCount; i++ {
		startingIndex := i * batchSize
		endingIndex := (i + 1) * batchSize
		if endingIndex > len(users) {
			endingIndex = len(users)
		}
		r[i] = users[startingIndex:endingIndex]
	}
	return r
}

func ChunkNewUserCalls(chain int64, users []contract_helpers.UserToken) [][]contract_helpers.UserToken {
	return ChunkArr[contract_helpers.UserToken](users, conf.MulticallMaxSize(chain))
}
