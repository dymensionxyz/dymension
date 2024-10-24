package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (k Keeper) validateBondDenom(ctx sdk.Context, c sdk.Coin) error {
	minBond := k.GetParams(ctx).MinBond
	if c.Denom != minBond.Denom {
		return errorsmod.Wrapf(types.ErrInvalidDenom, "expect: %s", minBond.Denom)
	}
	return nil
}

func (k Keeper) sufficientBond(ctx sdk.Context, c sdk.Coin) error {
	if err := k.validateBondDenom(ctx, c); err != nil {
		return err
	}
	minBond := k.GetParams(ctx).MinBond
	if !minBond.IsLTE(c) {
		return errorsmod.Wrapf(types.ErrInsufficientBond, "min: %s", minBond.Amount)
	}
	return nil
}

func (k Keeper) burn(ctx sdk.Context, seq *types.Sequencer, amt sdk.Coin) error {
	seq.SetTokensCoin(seq.TokensCoin().Sub(amt))
	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amt))
}

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
