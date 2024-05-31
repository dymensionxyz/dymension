package utilsmap

import (
	"cmp"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"golang.org/x/exp/constraints"
)

// SortedKeys returns the comparable sorted keys of a map
// Useful for deterministic iteration
func SortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

// Merge deterministically merges maps without allowing key conflicts
func Merge[M ~map[K]V, K constraints.Ordered, V any](maps ...M) (M, error) {
	ret := make(M)

	for _, m := range maps {
		for _, k := range SortedKeys(m) {
			if _, ok := ret[k]; ok {
				return nil, errorsmod.Wrapf(sdkerrors.ErrConflict, "duplicate key: %s", k)
			}
			ret[k] = m[k]
		}
	}

	return ret, nil
}
