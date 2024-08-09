package types

import (
	errorsmod "cosmossdk.io/errors"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	// TypeName is an alias variable of AssetType_AT_DYM_NAME
	TypeName = AssetType_AT_DYM_NAME

	// TypeAlias is an alias variable of AssetType_AT_ALIAS
	TypeAlias = AssetType_AT_ALIAS
)

var assetTypeFriendlyStrings = map[AssetType]string{
	AssetType_AT_DYM_NAME: "Dym-Name",
	AssetType_AT_ALIAS:    "Alias",
}

func (x AssetType) FriendlyString() string {
	if s, ok := assetTypeFriendlyStrings[x]; ok {
		return s
	}
	return "Unknown"
}

func ValidateOrderParams(params []string, assetType AssetType) error {
	switch assetType {
	case AssetType_AT_DYM_NAME:
		if len(params) != 0 {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"not accept order params for asset type: %s", assetType.FriendlyString(),
			)
		}
		return nil
	case AssetType_AT_ALIAS:
		if len(params) != 1 {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"expect 1 order param of RollApp ID for asset type: %s", assetType.FriendlyString(),
			)
		}
		if !dymnsutils.IsValidChainIdFormat(params[0]) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"invalid RollApp ID format: %s", params[0],
			)
		}
		return nil
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"unknown asset type: %s", assetType,
		)
	}
}
