package keeper

import (
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

// UnbondAllMatureSequencers unbonds all the mature unbonding sequencers that
// have finished their unbonding period.
func (k Keeper) UnbondAllMatureSequencers(ctx sdk.Context, currTime time.Time) {
	sequencers := k.GetMatureUnbondingSequencers(ctx, currTime)
	for _, seq := range sequencers {
		wrapFn := func(ctx sdk.Context) error {
			return k.unbondSequencer(ctx, seq.SequencerAddress)
		}
		err := osmoutils.ApplyFuncIfNoError(ctx, wrapFn)
		if err != nil {
			k.Logger(ctx).Error("failed to unbond sequencer", "error", err, "sequencer", seq.SequencerAddress)
			continue
		}
	}
}

// unbondSequencer unbonds a sequencer
func (k Keeper) unbondSequencer(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status != types.Unbonding {
		return sdkerrors.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is not unbonding: got %s",
			seq.Status.String(),
		)
	}

	if !seq.Tokens.Empty() {
		seqAcc, err := sdk.AccAddressFromBech32(seq.SequencerAddress)
		if err != nil {
			return err
		}

		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, seq.Tokens)
		if err != nil {
			return err
		}
	} else {
		k.Logger(ctx).Error("sequencer has no tokens to unbond", "sequencer", seq.SequencerAddress)
	}

	// set the status to unbonded and remove from the unbonding queue
	seq.Status = types.Unbonded
	seq.Tokens = sdk.Coins{}

	k.UpdateSequencer(ctx, seq, types.Unbonding)
	k.removeUnbondingSequencer(ctx, seq)

	return nil
}
