package v4

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func migrateSequencerIndices(ctx sdk.Context, k *sequencerkeeper.Keeper) error {
	list := k.AllSequencers(ctx)
	for _, oldSequencer := range list {

		// fill proposer index
		if oldSequencer.Proposer {
			k.SetProposer(ctx, oldSequencer.RollappId, oldSequencer.Address)
		}
		k.SetSuccessor(ctx, oldSequencer.RollappId, types.SentinelSeqAddr)

		// fill dymint proposer addr index
		addr, err := oldSequencer.ProposerAddr()
		if err != nil {
			// This shouldn't happen, but it's not obvious how we can recover from it.
			// It could lead to broken state for this rollapp, meaning that their IBC won't work properly.
			return errorsmod.Wrapf(err, "get dymint proposer address, seq: %s", oldSequencer.Address)
		}
		if err = k.SetSequencerByDymintAddr(ctx, addr, oldSequencer.Address); err != nil {
			return errorsmod.Wrapf(err, "set sequencer by dymint address: seq: %s", oldSequencer.Address)
		}

		// NOTE: technically should delete the unbonding queue, but we make an assumption
		// that the unbonding queue is empty at the time of upgrade.
	}
	return nil
}
