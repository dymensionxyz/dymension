package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CancelBuyOrder is message handler,
// handles canceling a Buy-Order, performed by the buyer who placed the offer.
func (k msgServer) CancelBuyOrder(goCtx context.Context, msg *dymnstypes.MsgCancelBuyOrder) (*dymnstypes.MsgCancelBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// get the Buy-Order record from store

	bo := k.GetBuyOrder(ctx, msg.OrderId)
	if bo == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order ID: %s", msg.OrderId)
	}

	var resp *dymnstypes.MsgCancelBuyOrderResponse
	var err error

	// process the Buy-Order based on the asset type

	if bo.AssetType == dymnstypes.TypeName || bo.AssetType == dymnstypes.TypeAlias {
		resp, err = k.processCancelBuyOrder(ctx, msg, *bo)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", bo.AssetType)
	}
	if err != nil {
		return nil, err
	}

	// charge protocol fee
	consumeMinimumGas(ctx, dymnstypes.OpGasCloseBuyOrder, "CancelBuyOrder")

	return resp, nil
}

// processCancelBuyOrder handles the message handled by CancelBuyOrder, type Dym-Name/Alias.
func (k msgServer) processCancelBuyOrder(
	ctx sdk.Context,
	msg *dymnstypes.MsgCancelBuyOrder, offer dymnstypes.BuyOrder,
) (*dymnstypes.MsgCancelBuyOrderResponse, error) {
	if err := k.validateCancelBuyOrder(ctx, msg, offer); err != nil {
		return nil, err
	}

	if err := k.RefundBuyOrder(ctx, offer); err != nil {
		return nil, err
	}

	if err := k.removeBuyOrder(ctx, offer); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgCancelBuyOrderResponse{}, nil
}

// validateCancelBuyOrder handles validation for the message handled by CancelBuyOrder, type Dym-Name/Alias.
func (k msgServer) validateCancelBuyOrder(_ sdk.Context, msg *dymnstypes.MsgCancelBuyOrder, offer dymnstypes.BuyOrder) error {
	if offer.Buyer != msg.Buyer {
		return errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the offer")
	}

	return nil
}

// removeBuyOrder removes the Buy-Order from the store and the reverse mappings.
func (k msgServer) removeBuyOrder(ctx sdk.Context, offer dymnstypes.BuyOrder) error {
	k.DeleteBuyOrder(ctx, offer.Id)

	err := k.RemoveReverseMappingBuyerToBuyOrder(ctx, offer.Buyer, offer.Id)
	if err != nil {
		return err
	}

	err = k.RemoveReverseMappingAssetIdToBuyOrder(ctx, offer.AssetId, offer.AssetType, offer.Id)
	if err != nil {
		return err
	}

	return nil
}
