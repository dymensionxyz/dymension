package keeper

import (
	"context"
	"time"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (k msgServer) PutAdsSellName(goCtx context.Context, msg *dymnstypes.MsgPutAdsSellName) (*dymnstypes.MsgPutAdsSellNameResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, params, err := k.validatePutAdsSellName(ctx, msg)
	if err != nil {
		return nil, err
	}

	so := msg.ToSellOrder()
	so.ExpireAt = ctx.BlockTime().Add(params.Misc.SellOrderDuration).Unix()

	if err := so.Validate(); err != nil {
		panic(errors.Wrap(err, "un-expected invalid state of created SO"))
	}

	if dymName.IsProhibitedTradingAt(time.Unix(so.ExpireAt, 0), params.Misc.ProhibitSellDuration) {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"%s before Dym-Name expiry, can not sell",
			params.Misc.ProhibitSellDuration,
		)
	}

	if err := k.SetSellOrder(ctx, so); err != nil {
		return nil, err
	}

	apoe := k.GetActiveSellOrdersExpiration(ctx)
	apoe.Add(so.Name, so.ExpireAt)
	if err := k.SetActiveSellOrdersExpiration(ctx, apoe); err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasPutAds, "PutAdsSellName")

	return &dymnstypes.MsgPutAdsSellNameResponse{}, nil
}

func (k msgServer) validatePutAdsSellName(ctx sdk.Context, msg *dymnstypes.MsgPutAdsSellName) (*dymnstypes.DymName, *dymnstypes.Params, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, nil, err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return nil, nil, dymnstypes.ErrDymNameNotFound.Wrap(msg.Name)
	}

	if dymName.Owner != msg.Owner {
		return nil, nil, sdkerrors.ErrUnauthorized
	}

	if dymName.IsExpiredAtCtx(ctx) {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrap("Dym-Name is already expired")
	}

	existingActiveSo := k.GetSellOrder(ctx, dymName.Name)
	if existingActiveSo != nil {
		if existingActiveSo.HasFinishedAtCtx(ctx) {
			return nil, nil, sdkerrors.ErrConflict.Wrap(
				"an active expired/completed Sell-Order already exists for the Dym-Name, must wait until processed",
			)
		}
		return nil, nil, sdkerrors.ErrConflict.Wrap("an active Sell-Order already exists for the Dym-Name")
	}

	params := k.GetParams(ctx)

	if msg.MinPrice.Denom != params.Price.PriceDenom {
		return nil, nil, sdkerrors.ErrInvalidRequest.Wrapf(
			"only %s is allowed as price",
			params.Price.PriceDenom,
		)
	}

	return dymName, &params, nil
}
