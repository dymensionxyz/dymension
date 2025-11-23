package collcompat

import (
	"bytes"
	"fmt"
	"sort"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"

	"github.com/cosmos/cosmos-sdk/types/kv"
)

// prefixInfo holds the prefix and its corresponding value codec.
type prefixInfo struct {
	Prefix []byte
	Codec  collcodec.UntypedValueCodec
}

// NewStoreDecoderFuncFromCollectionsSchema creates a function that can decode and stringify
// two KV pairs (kvA and kvB) using the provided collections schema.
// This function is typically used for generating human-readable state diffs.
func NewStoreDecoderFuncFromCollectionsSchema(schema collections.Schema) func(kvA, kvB kv.Pair) string {
	colls := schema.ListCollections()
	
	// Use a slice and sort it by prefix length for reliable prefix matching (longest prefix first).
	// This prevents accidentally matching a shorter prefix (e.g., 'U') before a longer one (e.g., 'US').
	prefixList := make([]prefixInfo, len(colls))
	for i, coll := range colls {
		prefixList[i] = prefixInfo{
			Prefix: coll.GetPrefix(),
			Codec:  coll.ValueCodec(),
		}
	}

	// Sort by prefix length in descending order.
	sort.Slice(prefixList, func(i, j int) bool {
		return len(prefixList[i].Prefix) > len(prefixList[j].Prefix)
	})

	return func(kvA, kvB kv.Pair) string {
		for _, info := range prefixList {
			prefix := info.Prefix
			vc := info.Codec
			
			if bytes.HasPrefix(kvA.Key, prefix) {
				// Check if kvB shares the same prefix. If not, report the mismatch instead of panicking.
				if !bytes.HasPrefix(kvB.Key, prefix) {
					return fmt.Sprintf("ERROR: prefix mismatch. Key A has prefix %X (%s), but Key B does not have it: %X (%s)",
						prefix, prefix, kvB.Key, kvB.Key)
				}
				
				// Decode kvA.Value
				vA, err := vc.Decode(kvA.Value)
				if err != nil {
					return fmt.Sprintf("ERROR: failed to decode kvA value (%X): %v", kvA.Value, err)
				}

				// Decode kvB.Value
				vB, err := vc.Decode(kvB.Value)
				if err != nil {
					return fmt.Sprintf("ERROR: failed to decode kvB value (%X): %v", kvB.Value, err)
				}
				
				// Stringify kvA
				vAString, err := vc.Stringify(vA)
				if err != nil {
					return fmt.Sprintf("ERROR: failed to stringify kvA value: %v", err)
				}
				
				// Stringify kvB
				vBString, err := vc.Stringify(vB)
				if err != nil {
					return fmt.Sprintf("ERROR: failed to stringify kvB value: %v", err)
				}
				
				// Return the clean, stringified representation of both values.
				return vAString + "\n" + vBString
			}
		}
		
		// If no prefix matches, report the unexpected key instead of panicking.
		return fmt.Sprintf("ERROR: unexpected key encountered: %X (%s). Does not match any known collection prefix.", kvA.Key, kvA.Key)
	}
}
