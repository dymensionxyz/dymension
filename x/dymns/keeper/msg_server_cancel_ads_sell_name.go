package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CancelAdsSellName is message handler,
// handles canceling Sell-Order, performed by the owner.
// This will stop the advertisement and remove the Dym-Name from the market.
// Can only be performed if the Dym-Name is not in any offer.
func (k msgServer) CancelAdsSellName(goCtx context.Context, msg *dymnstypes.MsgCancelAdsSellName) (*dymnstypes.MsgCancelAdsSellNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.validateCancelAdsSellName(ctx, msg); err != nil {
		return nil, err
	}

	k.DeleteSellOrder(ctx, msg.Name)

	aSoe := k.GetActiveSellOrdersExpiration(ctx)
	aSoe.Remove(msg.Name)
	if err := k.SetActiveSellOrdersExpiration(ctx, aSoe); err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasCloseAds, "CancelAdsSellName")

	return &dymnstypes.MsgCancelAdsSellNameResponse{}, nil
}

// validateCancelAdsSellName handles validation for the message handled by CancelAdsSellName.
func (k msgServer) validateCancelAdsSellName(ctx sdk.Context, msg *dymnstypes.MsgCancelAdsSellName) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return dymnstypes.ErrDymNameNotFound.Wrap(msg.Name)
	}

	if dymName.Owner != msg.Owner {
		return sdkerrors.ErrUnauthorized.Wrap("not the owner of the dym name")
	}

	so := k.GetSellOrder(ctx, msg.Name)
	if so == nil {
		return dymnstypes.ErrSellOrderNotFound.Wrap(msg.Name)
	}

	if so.HasExpiredAtCtx(ctx) {
		return dymnstypes.ErrInvalidState.Wrap("cannot cancel an expired order")
	}

	if so.HighestBid != nil {
		return dymnstypes.ErrInvalidState.Wrap("cannot cancel once bid placed")
	}

	return nil
}
