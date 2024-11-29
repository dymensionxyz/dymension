package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// Rollapp must exist - returns base units, not atto. E.g. 100 dym not 10e18 adym
func (k *Keeper) MinBond(ctx sdk.Context, rollappID string) math.Int {
	ra := k.MustGetRollapp(ctx, rollappID)
	return math.NewIntFromUint64(ra.MinSequencerBond)
}

func (k *Keeper) validMinBond(ctx sdk.Context, x uint64) error {
	if min_ := k.GetParams(ctx).MinSequencerBondGlobal; x < min_ {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "min sequencer bond is less than global min sequencer bond: min: %d, got: %d", min_, x)
	}
	return nil
}
