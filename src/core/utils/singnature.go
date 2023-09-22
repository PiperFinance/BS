package utils

import "github.com/ethereum/go-ethereum/crypto"

// EventTopicSignature accepts something like "ItemSet(bytes32,bytes32)" and returns with a hash like 0xe79e73da417710ae99aa2088575580a60415d359acfad9cdd3382d59c80281d4
// [Source](https://goethereumbook.org/en/event-read/)
func EventTopicSignature(event string) string {
	eventSignature := []byte(event)
	hash := crypto.Keccak256Hash(eventSignature)
	return hash.Hex()
}
