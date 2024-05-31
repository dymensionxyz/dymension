package utilsmemo

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	utilsmap "github.com/dymensionxyz/dymension/v3/utils/map"
)

type Memo = map[string]any

// Merge will merge the new objects into the existing memo, returning an error for any key clashes
// other should be an object which can marshal to json
// TODO: not the prettiest or most efficient
func Merge(memo string, other any) (string, error) {
	asMap := make(map[string]any)

	err := json.Unmarshal([]byte(memo), &asMap)
	if err != nil {
		return "", errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "memo")
	}

	/*
		We assume other can marshal to json and then we unmarshal it again to a map
	*/
	bz, err := json.Marshal(other)
	if err != nil {
		return "", errorsmod.Wrap(sdkerrors.ErrJSONMarshal, "other")
	}
	asMapOther := make(map[string]any)
	err = json.Unmarshal(bz, &asMapOther)
	if err != nil {
		return "", errorsmod.Wrap(sdkerrors.ErrJSONUnmarshal, "memo")
	}

	merged, err := utilsmap.Merge(asMap, asMapOther)
	if err != nil {
		return "", errorsmod.Wrapf(err, "map merge")
	}

	bz, err = json.Marshal(merged)
	if err != nil {
		return "", errorsmod.Wrap(sdkerrors.ErrJSONMarshal, "merged")
	}
	return string(bz), nil
}
