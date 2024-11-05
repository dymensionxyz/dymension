package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) bondDenom(ctx sdk.Context) string {
	return k.GetParams(ctx).MinBond.Denom
}

func (k Keeper) validBondDenom(ctx sdk.Context, c sdk.Coin) error {
	d := k.bondDenom(ctx)
	if c.Denom != d {
		return errorsmod.Wrapf(types.ErrInvalidDenom, "expect: %s", d)
	}
	return nil
}

func (k Keeper) sufficientBond(ctx sdk.Context, c sdk.Coin) error {
	if err := k.validBondDenom(ctx, c); err != nil {
		return err
	}
	minBond := k.GetParams(ctx).MinBond
	if c.IsLT(minBond) {
		return errorsmod.Wrapf(types.ErrInsufficientBond, "min: %s", minBond.Amount)
	}
	return nil
}

func (k Keeper) Kickable(ctx sdk.Context, proposer types.Sequencer) bool {
	kickThreshold := k.GetParams(ctx).KickThreshold
	return !proposer.Sentinel() && proposer.TokensCoin().IsLTE(kickThreshold)
}

func (k Keeper) burn(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	seq.SetTokensCoin(seq.TokensCoin().Sub(amt))
	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amt))
}

// Refund reduces the sequencer token balance by amt and refunds amt to the addr
func (k Keeper) refund(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	return errorsmod.Wrap(k.sendFromModule(ctx, seq, amt, seq.AccAddr()), "send tokens")
}

func (k Keeper) sendFromModule(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin, recipient sdk.AccAddress) error {
	seq.SetTokensCoin(seq.TokensCoin().Sub(amt))
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, sdk.NewCoins(amt))
}

func (k Keeper) sendToModule(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	seq.SetTokensCoin(seq.TokensCoin().Add(amt))
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, seq.AccAddr(), types.ModuleName, sdk.NewCoins(amt))
}
