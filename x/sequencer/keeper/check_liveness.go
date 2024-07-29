package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type CalculateNextSlashOrJailHeight func(
	HubBlockInterval time.Duration, // average time between hub blocks
	SlashTimeNoUpdate time.Duration, // time until first slash if not updating
	SlashInterval time.Duration, // gap between slash if still not updating
	JailTime time.Duration, // time until jail if not updating
	HubHeight int64, // current hub height
	LastRollappUpdateHeight int64, // when was the rollapp last updated
) (
	hubHeight int64, // height to schedule event
	isJail bool, // is it a jail event? (false -> slash)
)

func MulCoinsDec(coins sdk.Coins, dec sdk.Dec) sdk.Coins {
	return sdk.Coins{}
}

func (k Keeper) Slash(ctx sdk.Context, seq types.Sequencer, amt sdk.Coins, recipientAddr *sdk.AccAddress) error {
	seq.Tokens.Sub(amt...)
	if recipientAddr != nil {
		err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, *recipientAddr, amt)
		if err != nil {
			return errorsmod.Wrap(err, "send coins from module to account")
		}
	} else {
		err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, amt)
		if err != nil {
			return errorsmod.Wrap(err, "burn coins")
		}
	}
	// TODO: write back sequencer?
	return nil
}

func (k Keeper) Jail(ctx sdk.Context, seq types.Sequencer) error {
	// TODO: check contents of this, since it was copied

	oldStatus := seq.Status
	wasProposer := seq.Proposer
	// in case we are slashing an unbonding sequencer, we need to remove it from the unbonding queue
	if oldStatus == types.Unbonding {
		k.removeUnbondingSequencer(ctx, seq)
	}

	// set the status to unbonded
	seq.Status = types.Unbonded
	seq.Jailed = true
	seq.Proposer = false
	seq.UnbondingHeight = ctx.BlockHeight()
	seq.UnbondTime = ctx.BlockHeader().Time
	k.UpdateSequencer(ctx, seq, oldStatus)
	// rotate proposer if the slashed sequencer was the proposer
	if wasProposer {
		k.RotateProposer(ctx, seq.RollappId)
	}

	return nil
}
