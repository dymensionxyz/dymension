package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CancelSellOrder is message handler,
// handles canceling Sell-Order, performed by the owner.
// This will stop the advertisement and remove the Dym-Name/Alias sale from the market.
// Can only be performed if no one has placed a bid on the asset.
func (k msgServer) CancelSellOrder(goCtx context.Context, msg *dymnstypes.MsgCancelSellOrder) (*dymnstypes.MsgCancelSellOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	originalConsumedGas := ctx.GasMeter().GasConsumed()

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	var resp *dymnstypes.MsgCancelSellOrderResponse
	var err error

	// process the Sell-Order based on the asset type

	if msg.AssetType == dymnstypes.TypeName {
		resp, err = k.processCancelSellOrderWithAssetTypeDymName(ctx, msg)
	} else if msg.AssetType == dymnstypes.TypeAlias {
		resp, err = k.processCancelSellOrderWithAssetTypeAlias(ctx, msg)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", msg.AssetType)
	}
	if err != nil {
		return nil, err
	}

	// charge protocol fee
	consumeMinimumGas(ctx, dymnstypes.OpGasCloseSellOrder, originalConsumedGas, "CancelSellOrder")

	return resp, nil
}

// processCancelSellOrderWithAssetTypeDymName handles the message handled by CancelSellOrder, type Dym-Name.
func (k msgServer) processCancelSellOrderWithAssetTypeDymName(
	ctx sdk.Context, msg *dymnstypes.MsgCancelSellOrder,
) (*dymnstypes.MsgCancelSellOrderResponse, error) {
	if err := k.validateCancelSellOrderWithAssetTypeDymName(ctx, msg); err != nil {
		return nil, err
	}

	k.DeleteSellOrder(ctx, msg.AssetId, msg.AssetType)

	aSoe := k.GetActiveSellOrdersExpiration(ctx, msg.AssetType)
	aSoe.Remove(msg.AssetId)
	if err := k.SetActiveSellOrdersExpiration(ctx, aSoe, msg.AssetType); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCancelSellOrderResponse{}, nil
}

// validateCancelSellOrderWithAssetTypeDymName handles validation for the message handled by CancelSellOrder, type Dym-Name.
func (k msgServer) validateCancelSellOrderWithAssetTypeDymName(
	ctx sdk.Context, msg *dymnstypes.MsgCancelSellOrder,
) error {
	dymName := k.GetDymName(ctx, msg.AssetId)
	if dymName == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.AssetId)
	}

	if dymName.Owner != msg.Owner {
		return errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	so := k.GetSellOrder(ctx, msg.AssetId, msg.AssetType)
	if so == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", msg.AssetId)
	}

	if so.HasExpiredAtCtx(ctx) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot cancel an expired order")
	}

	if so.HighestBid != nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot cancel once bid placed")
	}

	return nil
}

// processCancelSellOrderWithAssetTypeAlias handles the message handled by CancelSellOrder, type Alias.
func (k msgServer) processCancelSellOrderWithAssetTypeAlias(
	ctx sdk.Context, msg *dymnstypes.MsgCancelSellOrder,
) (*dymnstypes.MsgCancelSellOrderResponse, error) {
	if err := k.validateCancelSellOrderWithAssetTypeAlias(ctx, msg); err != nil {
		return nil, err
	}

	k.DeleteSellOrder(ctx, msg.AssetId, msg.AssetType)

	aSoe := k.GetActiveSellOrdersExpiration(ctx, msg.AssetType)
	aSoe.Remove(msg.AssetId)
	if err := k.SetActiveSellOrdersExpiration(ctx, aSoe, msg.AssetType); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCancelSellOrderResponse{}, nil
}

// validateCancelSellOrderWithAssetTypeAlias handles validation for the message handled by CancelSellOrder, type Alias.
func (k msgServer) validateCancelSellOrderWithAssetTypeAlias(
	ctx sdk.Context, msg *dymnstypes.MsgCancelSellOrder,
) error {
	existingRollAppIdUsingAlias, found := k.GetRollAppIdByAlias(ctx, msg.AssetId)
	if !found {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "alias is not in-used: %s", msg.AssetId)
	}

	if !k.IsRollAppCreator(ctx, existingRollAppIdUsingAlias, msg.Owner) {
		return errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the RollApp")
	}

	so := k.GetSellOrder(ctx, msg.AssetId, msg.AssetType)
	if so == nil {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "Sell-Order: %s", msg.AssetId)
	}

	if so.HasExpiredAtCtx(ctx) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot cancel an expired order")
	}

	if so.HighestBid != nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "cannot cancel once bid placed")
	}

	return nil
}
