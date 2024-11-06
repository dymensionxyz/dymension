package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
)

func migrateSequencers(ctx sdk.Context, k *sequencerkeeper.Keeper) {
	list := k.AllSequencers(ctx)
	for _, oldSequencer := range list {
		newSequencer := ConvertOldSequencerToNew(oldSequencer)
		k.SetSequencer(ctx, newSequencer)
	}
}
