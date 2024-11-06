package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func migrateSequencerParams(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	p := sequencertypes.DefaultParams()

	// min bond is the only one that hasn't changed
	p.MinBond = k.GetParams(ctx).MinBond

	k.SetParams(ctx, p)
}
