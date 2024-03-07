package keeper

import "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

// combineKeys combines the byte arrays of multiple keys into a single byte array.
func combineKeys(keys ...[]byte) []byte {
	combined := []byte{}
	for i, key := range keys {
		combined = append(combined, key...)
		if i < len(keys)-1 { // not last item
			combined = append(combined, types.KeyIndexSeparator...)
		}
	}
	return combined
}
