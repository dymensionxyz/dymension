package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) MinBond(ctx sdk.Context, rollappID string) math.Int {
	return math.NewInt(0)
}
