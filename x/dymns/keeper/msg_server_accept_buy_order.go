package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// AcceptBuyOrder is message handler,
// handles accepting a Buy-Order or raising the amount for negotiation,
// performed by the owner of the asset.
func (k msgServer) AcceptBuyOrder(goCtx context.Context, msg *dymnstypes.MsgAcceptBuyOrder) (*dymnstypes.MsgAcceptBuyOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	originalConsumedGas := ctx.GasMeter().GasConsumed()

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// get the Buy-Order record from store

	bo := k.GetBuyOrder(ctx, msg.OrderId)
	if bo == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Buy-Order: %s", msg.OrderId)
	}

	miscParams := k.MiscParams(ctx)

	var resp *dymnstypes.MsgAcceptBuyOrderResponse
	var err error

	// process the Buy-Order based on the asset type

	if bo.AssetType == dymnstypes.TypeName {
		resp, err = k.processAcceptBuyOrderWithAssetTypeDymName(ctx, msg, *bo, miscParams)
	} else if bo.AssetType == dymnstypes.TypeAlias {
		resp, err = k.processAcceptBuyOrderWithAssetTypeAlias(ctx, msg, *bo, miscParams)
	} else {
		err = errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", bo.AssetType)
	}
	if err != nil {
		return nil, err
	}

	// charge protocol fee
	consumeMinimumGas(ctx, dymnstypes.OpGasUpdateBuyOrder, originalConsumedGas, "AcceptBuyOrder")

	return resp, nil
}

// processAcceptBuyOrderWithAssetTypeDymName handles the message handled by AcceptBuyOrder, type Dym-Name.
func (k msgServer) processAcceptBuyOrderWithAssetTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgAcceptBuyOrder, offer dymnstypes.BuyOrder, miscParams dymnstypes.MiscParams,
) (*dymnstypes.MsgAcceptBuyOrderResponse, error) {
	if !miscParams.EnableTradingName {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "trading of Dym-Name is disabled")
	}

	dymName, err := k.validateAcceptBuyOrderWithAssetTypeDymName(ctx, msg, offer)
	if err != nil {
		return nil, err
	}

	var accepted bool

	if msg.MinAccept.IsLT(offer.OfferPrice) {
		// this was checked earlier so this won't happen,
		// but I keep this here to easier to understand of all-cases of comparison
		panic("min-accept is less than offer price")
	} else if msg.MinAccept.IsEqual(offer.OfferPrice) {
		accepted = true

		// check active SO
		sellOrder := k.GetSellOrder(ctx, offer.AssetId, offer.AssetType)
		if sellOrder != nil {
			return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "must cancel the sell order first")
		}

		// take the offer
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			dymnstypes.ModuleName,
			sdk.MustAccAddressFromBech32(dymName.Owner),
			sdk.Coins{offer.OfferPrice},
		); err != nil {
			return nil, err
		}

		if err := k.removeBuyOrder(ctx, offer); err != nil {
			return nil, err
		}

		if err := k.transferDymNameOwnership(ctx, *dymName, offer.Buyer); err != nil {
			return nil, err
		}
	} else {
		accepted = false

		offer.CounterpartyOfferPrice = &msg.MinAccept
		if err := k.SetBuyOrder(ctx, offer); err != nil {
			return nil, err
		}
	}

	return &dymnstypes.MsgAcceptBuyOrderResponse{
		Accepted: accepted,
	}, nil
}

// validateAcceptBuyOrderWithAssetTypeDymName handles validation for the message handled by AcceptBuyOrder, type Dym-Name.
func (k msgServer) validateAcceptBuyOrderWithAssetTypeDymName(
	ctx sdk.Context,
	msg *dymnstypes.MsgAcceptBuyOrder, bo dymnstypes.BuyOrder,
) (*dymnstypes.DymName, error) {
	dymName := k.GetDymNameWithExpirationCheck(ctx, bo.AssetId)
	if dymName == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", bo.AssetId)
	}

	if dymName.Owner != msg.Owner {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	if bo.Buyer == msg.Owner {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "cannot accept own offer")
	}

	if msg.MinAccept.Denom != bo.OfferPrice.Denom {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"denom must be the same as the offer price: %s", bo.OfferPrice.Denom,
		)
	}

	if msg.MinAccept.IsLT(bo.OfferPrice) {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"amount must be greater than or equals to the offer price: %s", bo.OfferPrice.Denom,
		)
	}

	return dymName, nil
}

// processAcceptBuyOrderWithAssetTypeAlias handles the message handled by AcceptBuyOrder, type Alias.
func (k msgServer) processAcceptBuyOrderWithAssetTypeAlias(
	ctx sdk.Context,
	msg *dymnstypes.MsgAcceptBuyOrder, offer dymnstypes.BuyOrder, miscParams dymnstypes.MiscParams,
) (*dymnstypes.MsgAcceptBuyOrderResponse, error) {
	if !miscParams.EnableTradingAlias {
		return nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "trading of Alias is disabled")
	}

	if k.IsAliasPresentsInParamsAsAliasOrChainId(ctx, offer.AssetId) {
		// Please read the `processCompleteSellOrderWithAssetTypeAlias` method (msg_server_complete_sell_order.go) for more information.

		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied,
			"prohibited to trade aliases which is reserved for chain-id or alias in module params: %s", offer.AssetId,
		)
	}

	existingRollAppUsingAlias, err := k.validateAcceptBuyOrderWithAssetTypeAlias(ctx, msg, offer)
	if err != nil {
		return nil, err
	}

	destinationRollAppId := offer.Params[0]
	if !k.IsRollAppId(ctx, destinationRollAppId) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid destination Roll-App ID: %s", destinationRollAppId)
	}

	var accepted bool

	if msg.MinAccept.IsLT(offer.OfferPrice) {
		// this was checked earlier so this won't happen,
		// but I keep this here to easier to understand of all-cases of comparison
		panic("min-accept is less than offer price")
	} else if msg.MinAccept.IsEqual(offer.OfferPrice) {
		accepted = true

		// check active SO
		sellOrder := k.GetSellOrder(ctx, offer.AssetId, offer.AssetType)
		if sellOrder != nil {
			return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "must cancel the sell order first")
		}

		// take the offer
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(
			ctx,
			dymnstypes.ModuleName,
			sdk.MustAccAddressFromBech32(existingRollAppUsingAlias.Owner),
			sdk.Coins{offer.OfferPrice},
		); err != nil {
			return nil, err
		}

		if err := k.removeBuyOrder(ctx, offer); err != nil {
			return nil, err
		}

		if err := k.MoveAliasToRollAppId(ctx,
			existingRollAppUsingAlias.RollappId, // source Roll-App ID
			offer.AssetId,                       // alias
			destinationRollAppId,                // destination Roll-App ID
		); err != nil {
			return nil, err
		}
	} else {
		accepted = false

		offer.CounterpartyOfferPrice = &msg.MinAccept
		if err := k.SetBuyOrder(ctx, offer); err != nil {
			return nil, err
		}
	}

	return &dymnstypes.MsgAcceptBuyOrderResponse{
		Accepted: accepted,
	}, nil
}

// validateAcceptBuyOrderWithAssetTypeAlias handles validation for the message handled by AcceptBuyOrder, type Alias.
func (k msgServer) validateAcceptBuyOrderWithAssetTypeAlias(
	ctx sdk.Context,
	msg *dymnstypes.MsgAcceptBuyOrder, bo dymnstypes.BuyOrder,
) (*rollapptypes.Rollapp, error) {
	existingRollAppIdUsingAlias, found := k.GetRollAppIdByAlias(ctx, bo.AssetId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "alias is not in-used: %s", bo.AssetId)
	}

	if !k.IsRollAppCreator(ctx, existingRollAppIdUsingAlias, msg.Owner) {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the RollApp")
	}

	existingRollAppUsingAlias, found := k.rollappKeeper.GetRollapp(ctx, existingRollAppIdUsingAlias)
	if !found {
		// this can not happen as the previous check already ensures the Roll-App exists
		panic("roll-app not found: " + existingRollAppIdUsingAlias)
	}

	if bo.Buyer == msg.Owner {
		return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied, "cannot accept own offer")
	}

	if msg.MinAccept.Denom != bo.OfferPrice.Denom {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"denom must be the same as the offer price: %s", bo.OfferPrice.Denom,
		)
	}

	if msg.MinAccept.IsLT(bo.OfferPrice) {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"amount must be greater than or equals to the offer price: %s", bo.OfferPrice.Denom,
		)
	}

	return &existingRollAppUsingAlias, nil
}
