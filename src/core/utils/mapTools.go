package utils

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func SortedKeys[K constraints.Ordered, V any](anyMap map[K]V) []K {
	if len(anyMap) == 0 {
		return make([]K, 0)
	}
	keys := make([]K, 0, len(anyMap))
	for k := range anyMap {
		keys = append(keys, k)
	}
	less := func(i, j int) bool {
		return (keys[i] < keys[j])
	}
	sort.Slice(keys, less)
	return keys
}
