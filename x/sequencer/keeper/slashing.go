package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Slashing slashes the sequencer for misbehaviour
// Slashing can occur on both Bonded and Unbonding sequencers
func (k Keeper) Slashing(ctx sdk.Context, seqAddr string) error {
	seq, err := k.unbondSequencerAndBurn(ctx, seqAddr)
	if err != nil {
		return fmt.Errorf("slash sequencer: %w", err)
	}

	seq.Jailed = true
	seq.UnbondRequestHeight = ctx.BlockHeight()
	seq.UnbondTime = ctx.BlockTime()
	k.UpdateSequencer(ctx, *seq)

	// emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeSlashed,
			sdk.NewAttribute(types.AttributeKeySequencer, seqAddr),
		),
	)

	return nil
}

// InstantUnbondAllSequencers unbonds all sequencers for a rollapp
// This is called by the `FraudSubmitted` hook
func (k Keeper) InstantUnbondAllSequencers(ctx sdk.Context, rollappID string) error {
	// unbond all bonded/unbonding sequencers
	bonded := k.GetSequencersByRollappByStatus(ctx, rollappID, types.Bonded)
	unbonding := k.GetSequencersByRollappByStatus(ctx, rollappID, types.Unbonding)
	for _, sequencer := range append(bonded, unbonding...) {
		err := k.unbondSequencer(ctx, sequencer.SequencerAddress)
		if err != nil {
			return err
		}
	}

	return nil
}
