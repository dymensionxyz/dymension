package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k *Keeper) MinBond(ctx sdk.Context, rollappID string) sdk.Coin {
	ra := k.MustGetRollapp(ctx, rollappID)
	return ra.MinSequencerBond[0]
}

func (k *Keeper) validMinBond(ctx sdk.Context, x sdk.Coin) error {
	if err := types.ValidateBasicMinSeqBond(x); err != nil {
		return err
	}
	min_ := k.GetParams(ctx).MinSequencerBondGlobal
	if x.IsLT(min_) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min sequencer bond is less than global min sequencer bond: min: %s, got: %s", min_, x)
	}
	return nil
}
