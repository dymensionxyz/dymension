package utilsmemo

import (
	"golang.org/x/exp/constraints"

	utilsmap "github.com/dymensionxyz/dymension/v3/utils/map"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Memo = map[string]any

// Merge will merge the new objects into the existing memo, returning an error for any key clashes
func Merge[M ~map[K]V, K constraints.Ordered, V any](maps ...M) (M, error) {
	ret := make(M)

	for _, m := range maps {
		for _, k := range utilsmap.SortedKeys(m) {
			if _, ok := ret[k]; ok {
				return nil, errorsmod.Wrapf(sdkerrors.ErrConflict, "duplicate key: %s", k)
			}
			ret[k] = m[k]
		}
	}

	return ret, nil
}
