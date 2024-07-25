package keeper

import (
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func MulCoinsDec(coins sdk.Coins, dec sdk.Dec) sdk.Coins {
	return sdk.Coins{}
}

func (k Keeper) LivenessSlashAndJail(ctx sdk.Context,
	hUpdate int64,
	hubBlockTime time.Duration,
	slashTimeNoUpdate time.Duration,
	slashTimeNoTerminalUpdate time.Duration,
	slashInterval time.Duration,
	slashMultiplier sdk.Dec,
	jailTime time.Duration,
	minBond sdk.Coins, // TODO: comes from where?
	seqAddr string,
	burnMultiplier sdk.Dec, recipients ...types.LivenessSlashAndJailFundsRecipient,
) (types.LivenessSlashAndJailResult, error) {
	seq, found := k.GetSequencer(ctx, seqAddr)
	if !found {
		return types.LivenessSlashAndJailResult{}, errorsmod.Wrap(gerrc.ErrNotFound, "get sequencer")
	}

	// TODO: check assumption that they aren't already jailed

	args := types.LivenessSlashAndJailArgs{
		HHub:                      ctx.BlockHeight(),
		HNoticeExpired:            nil, // TODO:
		HUpdate:                   hUpdate,
		HubBlockTime:              hubBlockTime,
		SlashTimeNoUpdate:         slashTimeNoUpdate,
		SlashTimeNoTerminalUpdate: slashTimeNoTerminalUpdate,
		SlashInterval:             slashInterval,
		SlashMultiplier:           slashMultiplier,
		JailTime:                  jailTime,
		Balance:                   seq.Tokens, // TODO: need to handle 0 case?
		MinBond:                   minBond,
	}

	slashAmt, jail := args.Calculate()

	if err := k.Slash(ctx, seq, MulCoinsDec(slashAmt, burnMultiplier), nil); err != nil {
		return types.LivenessSlashAndJailResult{}, err // TODO:
	}

	for _, r := range recipients {
		if err := k.Slash(ctx, seq, MulCoinsDec(slashAmt, r.Multiplier), &r.Addr); err != nil {
			return types.LivenessSlashAndJailResult{}, err // TODO:
		}
	}

	if jail {
		if err := k.Jail(ctx, seq); err != nil {
			return types.LivenessSlashAndJailResult{}, err // TODO:
		}
	}

	return types.LivenessSlashAndJailResult{
		Slashed:                    slashAmt,
		Jailed:                     jail,
		TimeUntilNextSlashPossible: time.Time{}, // TODO:
		FundsReceived:              nil,         // TODO:
	}, nil
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
