package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) UnbondSequencer(ctx sdk.Context, seqAddr string, jailed bool) (time.Time, error) {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return time.Time{}, types.ErrUnknownSequencer
	}

	if seq.Status != types.Bonded || seq.Status != types.Proposer {
		return time.Time{}, sdkerrors.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is not bonded: got %s",
			seq.Status.String(),
		)
	}

	// set the status to unbonding
	seq.Status = types.Unbonding
	seq.Jailed = jailed
	seq.UnbondingHeight = ctx.BlockHeight()

	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	seq.UnbondingTime = completionTime

	k.SetSequencer(ctx, seq)

	return completionTime, nil
}
