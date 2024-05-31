package utilsmemo

import (
	"encoding/json"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	utilsmap "github.com/dymensionxyz/dymension/v3/utils/map"
)

type Memo = map[string]any

// Merge will merge the new objects into the existing memo, returning an error for any key clashes
func Merge(memo string, new ...map[string]any) (string, error) {
	asMap := make(map[string]any)

	err := json.Unmarshal([]byte(memo), &asMap)
	if err != nil {
		return "", sdkerrors.ErrJSONUnmarshal
	}

	merged, err := utilsmap.Merge(asMap, new[0])

	bz, err := json.Marshal(merged)
	if err != nil {
		return "", sdkerrors.ErrJSONMarshal
	}
	return string(bz), nil
}
