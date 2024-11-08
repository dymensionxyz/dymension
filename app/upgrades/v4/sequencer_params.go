package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func migrateSequencerParams(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	// (Note: all of them have changed, including min bond)
	p := sequencertypes.DefaultParams()

	k.SetParams(ctx, p)
}
