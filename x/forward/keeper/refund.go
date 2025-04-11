package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (k Keeper) refundOnError(ctx sdk.Context, f func() error, srcAddr sdk.AccAddress, refundAddr sdk.AccAddress, coins sdk.Coins) {
	err := f()
	if err != nil {
		_ = ctx.EventManager().EmitTypedEvent(&types.EventWillRefund{
			ErrCause:   err.Error(),
			RefundAddr: refundAddr.String(),
		})

		errRefund := k.bankK.SendCoins(ctx, srcAddr, refundAddr, coins)

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
