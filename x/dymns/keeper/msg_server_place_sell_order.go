package keeper

import (
	"context"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// TODO DymNS: bidder should be Roll-App owner

// PlaceSellOrder is message handler,
// handles creating a Sell-Order that advertise a Dym-Name/Alias is for sale, performed by the owner.
func (k msgServer) PlaceSellOrder(goCtx context.Context, msg *dymnstypes.MsgPlaceSellOrder) (*dymnstypes.MsgPlaceSellOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	params := k.GetParams(ctx)

	var resp *dymnstypes.MsgPlaceSellOrderResponse
	var err error

	if msg.OrderType == dymnstypes.NameOrder {
		resp, err = k.processPlaceSellOrderTypeDymName(ctx, msg, params)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", msg.OrderType)
	}
	if err != nil {
		return nil, err
	}

	consumeMinimumGas(ctx, dymnstypes.OpGasPlaceSellOrder, "PlaceSellOrder")

	return resp, nil
}

// processPlaceSellOrderTypeDymName handles the message handled by PlaceSellOrder, type Dym-Name.
func (k msgServer) processPlaceSellOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceSellOrder, params dymnstypes.Params,
) (*dymnstypes.MsgPlaceSellOrderResponse, error) {
	dymName, err := k.validatePlaceSellOrderTypeDymName(ctx, msg, params)
	if err != nil {
		return nil, err
	}

	so := msg.ToSellOrder()
	so.ExpireAt = ctx.BlockTime().Add(params.Misc.SellOrderDuration).Unix()

	if err := so.Validate(); err != nil {
		panic(errorsmod.Wrap(err, "un-expected invalid state of created SO"))
	}

	if dymName.IsProhibitedTradingAt(time.Unix(so.ExpireAt, 0), params.Misc.ProhibitSellDuration) {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition,
			"duration before Dym-Name expiry, prohibited to sell: %s",
			params.Misc.ProhibitSellDuration,
		)
	}

	if err := k.SetSellOrder(ctx, so); err != nil {
		return nil, err
	}

	aSoe := k.GetActiveSellOrdersExpiration(ctx, so.Type)
	aSoe.Add(so.GoodsId, so.ExpireAt)
	if err := k.SetActiveSellOrdersExpiration(ctx, aSoe, so.Type); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgPlaceSellOrderResponse{}, nil
}

// validatePlaceSellOrderTypeDymName handles validation for message handled by PlaceSellOrder, type Dym-Name.
func (k msgServer) validatePlaceSellOrderTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgPlaceSellOrder, params dymnstypes.Params,
) (*dymnstypes.DymName, error) {
	dymName := k.GetDymName(ctx, msg.GoodsId)
	if dymName == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.GoodsId)
	}

	if dymName.Owner != msg.Owner {
		return nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	if dymName.IsExpiredAtCtx(ctx) {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "Dym-Name is already expired")
	}

	existingActiveSo := k.GetSellOrder(ctx, dymName.Name, msg.OrderType)
	if existingActiveSo != nil {
		if existingActiveSo.HasFinishedAtCtx(ctx) {
			return nil, errorsmod.Wrap(
				gerrc.ErrAlreadyExists,
				"an active expired/completed Sell-Order already exists for the Dym-Name, must wait until processed",
			)
		}
		return nil, errorsmod.Wrap(gerrc.ErrAlreadyExists, "an active Sell-Order already exists for the Dym-Name")
	}

	if msg.MinPrice.Denom != params.Price.PriceDenom {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"the only denom allowed as price: %s", params.Price.PriceDenom,
		)
	}

	return dymName, nil
}
