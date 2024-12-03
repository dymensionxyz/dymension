package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func IsUpdateMinSeqBond(c sdk.Coin) bool {
	return ValidateBasicMinSeqBond(c) == nil
}

func ValidateBasicMinSeqBond(c sdk.Coin) error {
	if err := c.Validate(); err != nil {
		// thanks coins api!
		return errorsmod.Wrap(err, "validate")
	}
	return ValidateBasicMinSeqBondCoins(sdk.NewCoins(c))
}

func ValidateBasicMinSeqBondCoins(c sdk.Coins) error {
	if err := c.Validate(); err != nil {
		return errorsmod.Wrap(err, "validate")
	}
	if c.Len() != 1 {
		return gerrc.ErrInvalidArgument.Wrap("not exactly one coin")
	}
	if !c.IsAllPositive() {
		return gerrc.ErrInvalidArgument.Wrap("non positive")
	}
	if c[0].Denom != commontypes.DYMCoin.Denom {
		return gerrc.ErrInvalidArgument.Wrap("denom")
	}
	return nil
}
