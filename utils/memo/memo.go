package utilsmemo

import (
	"encoding/json"

	utilsmap "github.com/dymensionxyz/dymension/v3/utils/map"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type Memo = map[string]any

// Merge will merge the new objects into the existing memo, returning an error for any key clashes
func Merge(memo string, new ...map[string]any) (string, error) {
	asMap := make(map[string]any)

	err := json.Unmarshal([]byte(memo), &asMap)
	if err != nil {
		return "", sdkerrors.ErrJSONUnmarshal
	}
	for _, newMemo := range new {
		for _, k := range utilsmap.SortedKeys(newMemo) {
			if _, ok := asMap[k]; ok {
				return "", errorsmod.Wrapf(sdkerrors.ErrConflict, "duplicate memo key: %s", k)
			}
			asMap[k] = newMemo[k]
		}
	}

	bz, err := json.Marshal(asMap)
	if err != nil {
		return "", sdkerrors.ErrJSONMarshal
	}
	return string(bz), nil
}
