package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Unbond defines a method for removing coins from sequencer's bond
func (k Keeper) Slashing(ctx sdk.Context, seqAddr string) error {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.ErrUnknownSequencer
	}

	if seq.Status == types.Unbonded {
		return sdkerrors.Wrap(
			types.ErrInvalidSequencerStatus,
			"cant slash unbonded sequencer",
		)
	}

	if seq.Tokens.IsPositive() {
		coins := sdk.NewCoins(seq.Tokens)
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coins)
		if err != nil {
			return err
		}

		seq.Tokens = sdk.Coin{}
	}

	oldStatus := seq.Status
	// set the status to unbonded
	seq.Status = types.Unbonded
	seq.Jailed = true
	seq.UnbondingHeight = ctx.BlockHeight()
	seq.UnbondTime = ctx.BlockHeader().Time

	k.UpdateSequencer(ctx, seq, oldStatus)

	return nil
}
