package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) StartUnbondingSequencer(ctx sdk.Context, seqAddr string) (time.Time, error) {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return time.Time{}, types.ErrUnknownSequencer
	}

	if seq.Status != types.Bonded && seq.Status != types.Proposer {
		return time.Time{}, sdkerrors.Wrapf(
			types.ErrInvalidSequencerStatus,
			"sequencer status is not bonded: got %s",
			seq.Status.String(),
		)
	}

	// set the status to unbonding
	seq.Status = types.Unbonding
	seq.UnbondingHeight = ctx.BlockHeight()

	completionTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))
	seq.UnbondingTime = completionTime

	k.SetSequencer(ctx, seq)

	//FIXME: set in some unbonding queue

	return completionTime, nil
}

// UnbondAllMatureSequencers unbonds all the mature unbonding sequencers that
// have finished their unbonding period.
func (k Keeper) UnbondAllMatureSequencers(ctx sdk.Context, currTime time.Time) {
	sequencers := k.GetUnbondingSequencers(ctx)
	for _, seq := range sequencers {
		if seq.UnbondingTime.Before(currTime) {
			//FIXME: wrap with applyIf
			err := k.unbondSequencer(ctx, seq.SequencerAddress)
			if err != nil {
				k.Logger(ctx).Error("failed to unbond sequencer", "error", err, "sequencer", seq.SequencerAddress)
				continue
			}
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

	if seq.Tokens.IsPositive() {
		seqAcc, err := sdk.AccAddressFromBech32(seq.SequencerAddress)
		if err != nil {
			return err
		}

		coins := sdk.NewCoins(*seq.Tokens)
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, seqAcc, coins)
		if err != nil {
			return err
		}
	}

	// set the status to unbonded
	seq.Status = types.Unbonded
	seq.Tokens = &sdk.Coin{}

	k.SetSequencer(ctx, seq)
	return nil
}
