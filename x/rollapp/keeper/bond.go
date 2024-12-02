package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k *Keeper) MinBond(ctx sdk.Context, rollappID string) sdk.Coin {
	ra := k.MustGetRollapp(ctx, rollappID)
	amt := math.NewIntFromUint64(ra.MinSequencerBond)
	return commontypes.ADym(amt)
}

func (k *Keeper) validMinBond(ctx sdk.Context, x uint64) error {
	if min_ := k.GetParams(ctx).MinSequencerBondGlobal; x < min_ {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min sequencer bond is less than global min sequencer bond: min: %d, got: %d", min_, x)
	}
	return nil
}
