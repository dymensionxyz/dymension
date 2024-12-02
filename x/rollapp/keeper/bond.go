package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k *Keeper) MinBond(ctx sdk.Context, rollappID string) sdk.Coin {
	ra := k.MustGetRollapp(ctx, rollappID)
	return ra.MinSequencerBond
}

func (k *Keeper) validMinBond(ctx sdk.Context, x sdk.Coin) error {
	min_ := k.GetParams(ctx).MinSequencerBondGlobal
	if x.Denom != min_.Denom {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min sequencer bond denom is not equal to global min sequencer bond denom: min: %s, got: %s", min_.Denom, x.Denom)
	}
	if x.IsLT(min_) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min sequencer bond is less than global min sequencer bond: min: %s, got: %s", min_, x)
	}
	return nil
}
