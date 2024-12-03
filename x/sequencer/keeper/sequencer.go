package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// SentinelSequencer is a convenient placeholder for the empty-sequencer case
// Note: does not populate rollappID by default
func (k Keeper) SentinelSequencer(ctx sdk.Context) types.Sequencer {
	s := k.NewSequencer(ctx, "")
	s.Status = types.Bonded
	s.Address = types.SentinelSeqAddr
	s.OptedIn = true
	return *s
}

func (k Keeper) NewSequencer(ctx sdk.Context, rollapp string) *types.Sequencer {
	return &types.Sequencer{
		RollappId: rollapp,
		// DO NOT USE NEW COINS! IT WILL REMOVE ZERO COIN
		Tokens: sdk.Coins{sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.NewInt(0))},
	}
}
