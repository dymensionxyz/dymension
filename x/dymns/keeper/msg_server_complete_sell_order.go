package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CompleteSellOrder is message handler,
// handles Sell-Order completion action, can be performed by either asset owner or the person who placed the highest bid.
// Can only be performed when Sell-Order expired and has a bid placed.
// If the asset was expired or prohibited trading, bid placed will be force to return to the bidder, ownership will not be transferred.
func (k msgServer) CompleteSellOrder(goCtx context.Context, msg *dymnstypes.MsgCompleteSellOrder) (*dymnstypes.MsgCompleteSellOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	originalConsumedGas := ctx.GasMeter().GasConsumed()

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	var resp *dymnstypes.MsgCompleteSellOrderResponse
	var err error

	// process the Sell-Order based on the asset type

	if msg.AssetType == dymnstypes.TypeName {
		resp, err = k.processCompleteSellOrderWithAssetTypeDymName(ctx, msg)
	} else if msg.AssetType == dymnstypes.TypeAlias {
		resp, err = k.processCompleteSellOrderWithAssetTypeAlias(ctx, msg)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", msg.AssetType)
	}
	if err != nil {
		return nil, err
	}

	// charge protocol fee
	consumeMinimumGas(ctx, dymnstypes.OpGasCompleteSellOrder, originalConsumedGas, "CompleteSellOrder")

	return resp, nil
}

// processCompleteSellOrderWithAssetTypeDymName handles the message handled by CompleteSellOrder, type Dym-Name.
func (k msgServer) processCompleteSellOrderWithAssetTypeDymName(
	ctx sdk.Context, msg *dymnstypes.MsgCompleteSellOrder,
) (*dymnstypes.MsgCompleteSellOrderResponse, error) {
	so, dymName, err := k.validateCompleteSellOrderWithAssetTypeDymName(ctx, msg)
	if err != nil {
		return nil, err
	}

	miscParams := k.MiscParams(ctx)

	var refund bool
	if !miscParams.EnableTradingName {
		k.Logger(ctx).Info("Dym-Name trading is disabled, refunding the bid", "Dym-Name", dymName.Name)
		refund = true
	} else if dymName.IsExpiredAtCtx(ctx) {
		k.Logger(ctx).Info("Dym-Name is expired, refunding the bid", "Dym-Name", dymName.Name)
		refund = true
	}

	if refund {
		if err := k.RefundBid(ctx, *so.HighestBid, so.AssetType); err != nil {
			return nil, err
		}
		k.DeleteSellOrder(ctx, so.AssetId, so.AssetType)
		return &dymnstypes.MsgCompleteSellOrderResponse{}, nil
	}

	if err := k.CompleteDymNameSellOrder(ctx, so.AssetId); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCompleteSellOrderResponse{}, nil
}

// validateCompleteSellOrderWithAssetTypeDymName handles validation for the message handled by CompleteSellOrder, type Dym-Name.
func (k msgServer) validateCompleteSellOrderWithAssetTypeDymName(
	ctx sdk.Context, msg *dymnstypes.MsgCompleteSellOrder,
) (*dymnstypes.SellOrder, *dymnstypes.DymName, error) {
	so := k.GetSellOrder(ctx, msg.AssetId, msg.AssetType)
	if so == nil {
		return nil, nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", msg.AssetId)
	}

	if so.HighestBid == nil {
		return nil, nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no bid placed on the Sell-Order")
	}

	if !so.HasFinishedAtCtx(ctx) {
		return nil, nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "Sell-Order not yet completed")
	}

	dymName := k.GetDymName(ctx, msg.AssetId)
	if dymName == nil {
		return nil, nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.AssetId)
	}

	if dymName.Owner != msg.Participant && so.HighestBid.Bidder != msg.Participant {
		return nil, nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "must be either Dym-Name owner or the highest bidder to complete the Sell-Order")
	}

	return so, dymName, nil
}

// processCompleteSellOrderWithAssetTypeAlias handles the message handled by CompleteSellOrder, type Alias.
func (k msgServer) processCompleteSellOrderWithAssetTypeAlias(
	ctx sdk.Context, msg *dymnstypes.MsgCompleteSellOrder,
) (*dymnstypes.MsgCompleteSellOrderResponse, error) {
	so, err := k.validateCompleteSellOrderWithAssetTypeAlias(ctx, msg)
	if err != nil {
		return nil, err
	}

	miscParams := k.MiscParams(ctx)
	var refund bool

	/**
	For the Sell-Orders which the assets are prohibited to trade,
	the Sell-Order will be force cancelled and the bids will be refunded.

	Why some aliases are prohibited to trade? And what are they?
	In module params, there is a list of alias mapping for some external well-known chains.
	So those aliases are considered as reserved for the external chains,
	therefor trading is not allowed.

	Why can someone own a prohibited alias?
	An alias can be bought before the reservation was made.
	But when the alias becomes reserved for the external well-known chains,
	the alias will be prohibited to trade.

	Why can someone place a Sell-Order for the prohibited alias?
	When a Sell-Order created before the reservation was made.
	*/
	if forceCancel := k.IsAliasPresentsInParamsAsAliasOrChainId(ctx, so.AssetId); forceCancel {
		// Sell-Order will be force cancelled and refund bids if any,
		// when the alias is prohibited to trade
		k.Logger(ctx).Info("Alias is prohibited to trade, refunding the bid", "Alias", so.AssetId)
		refund = true
	} else if !miscParams.EnableTradingAlias {
		k.Logger(ctx).Info("Alias trading is disabled, refunding the bid", "Alias", so.AssetId)
		refund = true
	}

	if refund {
		if err := k.RefundBid(ctx, *so.HighestBid, so.AssetType); err != nil {
			return nil, err
		}
		k.DeleteSellOrder(ctx, so.AssetId, so.AssetType)
		return &dymnstypes.MsgCompleteSellOrderResponse{}, nil
	}

	if err := k.CompleteAliasSellOrder(ctx, msg.AssetId); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCompleteSellOrderResponse{}, nil
}

// validateCompleteSellOrderWithAssetTypeAlias handles validation for the message handled by CompleteSellOrder, type Alias.
func (k msgServer) validateCompleteSellOrderWithAssetTypeAlias(
	ctx sdk.Context, msg *dymnstypes.MsgCompleteSellOrder,
) (*dymnstypes.SellOrder, error) {
	so := k.GetSellOrder(ctx, msg.AssetId, msg.AssetType)
	if so == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", msg.AssetId)
	}

	if so.HighestBid == nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no bid placed on the Sell-Order")
	}

	if !so.HasFinishedAtCtx(ctx) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "Sell-Order not yet completed")
	}

	existingRollAppIdUsingAlias, found := k.GetRollAppIdByAlias(ctx, msg.AssetId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "alias is not in-use: %s", msg.AssetId)
	}

	if !k.IsRollAppCreator(ctx, existingRollAppIdUsingAlias, msg.Participant) && so.HighestBid.Bidder != msg.Participant {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "must be either Roll-App creator or the highest bidder to complete the Sell-Order")
	}

	return so, nil
}
