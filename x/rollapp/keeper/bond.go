package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Rollapp must exist - returns base units, not atto. E.g. 100 dym not 10e18 adym
func (k *Keeper) MinBond(ctx sdk.Context, rollappID string) math.Int {
	ra := k.MustGetRollapp(ctx, rollappID)
	return math.NewIntFromUint64(ra.MinSequencerBond)
}
