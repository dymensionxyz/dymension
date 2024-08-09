package types

import (
	errorsmod "cosmossdk.io/errors"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	// NameOrder is an alias variable of OrderType_OT_DYM_NAME
	NameOrder = OrderType_OT_DYM_NAME

	// AliasOrder is an alias variable of OrderType_OT_ALIAS
	AliasOrder = OrderType_OT_ALIAS
)

var orderTypeFriendlyStrings = map[OrderType]string{
	OrderType_OT_DYM_NAME: "Dym-Name",
	OrderType_OT_ALIAS:    "Alias",
}

func (x OrderType) FriendlyString() string {
	if s, ok := orderTypeFriendlyStrings[x]; ok {
		return s
	}
	return "Unknown"
}

func OrderTypeFromFriendlyString(s string) OrderType {
	for k, v := range orderTypeFriendlyStrings {
		if v == s {
			return k
		}
	}
	return OrderType_OT_UNKNOWN
}

func ValidateOrderParams(params []string, orderType OrderType) error {
	switch orderType {
	case OrderType_OT_DYM_NAME:
		if len(params) != 0 {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"not accept order params for order type: %s", orderType.FriendlyString(),
			)
		}
		return nil
	case OrderType_OT_ALIAS:
		if len(params) != 1 {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
				"expect 1 order param of RollApp ID for order type: %s", orderType.FriendlyString(),
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
			"unknown order type: %s", orderType,
		)
	}
}
