package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) refundOnError(ctx sdk.Context, f func() error,
	srcAddr sdk.AccAddress,
	srcModule string,
	dstAddr sdk.AccAddress, coin sdk.Coin) {

	// avoid footguns
	if srcModule != "" && 0 < len(srcAddr) {
		panic("srcModule and srcAddr cannot both be set")
	}

	err := f()
	if err != nil {
		_ = ctx.EventManager().EmitTypedEvent(&types.EventWillRefund{
			ErrCause:   err.Error(),
			RefundAddr: dstAddr.String(),
		})

		var errRefund error
		if srcModule != "" {
			errRefund = k.bankK.SendCoinsFromModuleToAccount(ctx, srcModule, dstAddr, sdk.NewCoins(coin))
		} else {
			errRefund = k.bankK.SendCoins(ctx, srcAddr, dstAddr, sdk.NewCoins(coin))
		}

		if errRefund != nil {
			// should never happen
			errLog := types.RefundFail{
				Addr:      dstAddr.String(),
				Coins:     coin,
				ErrCause:  err,
				ErrRefund: errRefund,
			}
			k.Logger(ctx).Error("There was an error but refund failed.", "error", errLog)
		}
	}
}
