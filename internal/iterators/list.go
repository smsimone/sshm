package iterators

import (
	"cmp"
	"slices"
)

func GetMaxFunc[T any, K cmp.Ordered](seq []T, getKey func(T) K) K {
	keys := make([]K, len(seq))
	for idx, item := range seq {
		keys[idx] = getKey(item)
	}
	return slices.Max(keys)
}
