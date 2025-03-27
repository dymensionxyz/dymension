package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) escrowFromModule(ctx sdk.Context, srcModule string, c sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromModuleToModule(ctx, srcModule, types.ModuleName, c)
}

func (k Keeper) escrowFromUser(ctx sdk.Context, srcAcc sdk.AccAddress, c sdk.Coins) error {
	return k.bankKeeper.SendCoinsFromAccountToModule(ctx, srcAcc, types.ModuleName, c)
}

func (k Keeper) refundFromModule(ctx sdk.Context, dstAddr sdk.AccAddress, c sdk.Coins) error {
	// TODO: event
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, dstAddr, c)
}

func getRefundAddr(recovery types.Recovery) sdk.AccAddress {
	return recovery.MustAddr()
}

func (k Keeper) refundOnError(ctx sdk.Context, f func() error, r types.Recovery, coins sdk.Coins) {
	err := f()
	if err != nil {
		_ = ctx.EventManager().EmitTypedEvent(&types.EventWillRefund{
			ErrCause: err.Error(),
		})

		refundAddr := getRefundAddr(r)
		errRefund := k.refundFromModule(ctx, refundAddr, coins)
		if errRefund != nil {
			// should never happen
			errLog := types.RefundFail{
				Addr:      refundAddr.String(),
				Coins:     coins,
				ErrCause:  err,
				ErrRefund: errRefund,
			}
			k.Logger(ctx).Error("There was an error but refund failed.", "error", errLog)
		}
	}
}
