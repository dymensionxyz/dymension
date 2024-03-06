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

// findIndex takes an array of IDs. Then return the index of a specific ID.
func findIndex(IDs []uint64, ID uint64) int {
	for index, id := range IDs {
		if id == ID {
			return index
		}
	}
	return -1
}
