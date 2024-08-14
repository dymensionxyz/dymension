package keeper

import (
	"context"
	"errors"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// RegisterAlias is message handler, handles registration of a new Alias for an existing RollApp.
func (k msgServer) RegisterAlias(goCtx context.Context, msg *dymnstypes.MsgRegisterAlias) (*dymnstypes.MsgRegisterAliasResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := k.validateRegisterAlias(ctx, msg)
	if err != nil {
		return nil, err
	}

	moduleParams := k.GetParams(ctx)

	registrationCost := sdk.NewCoin(
		moduleParams.Price.PriceDenom, moduleParams.Price.GetAliasPrice(msg.Alias),
	)

	if !registrationCost.Equal(msg.ConfirmPayment) {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"actual payment is is different with provided by user: %s != %s", registrationCost.String(), msg.ConfirmPayment,
		)
	}

	// At this place we don't do compare actual payment with estimated payment calculated by EstimateRegisterAlias
	// because in-case there is different between them, it would prevent RollApp owners to registration.

	if err := k.registerAliasForRollApp(
		ctx,
		msg.RollappId, sdk.MustAccAddressFromBech32(msg.Owner),
		msg.Alias,
		sdk.NewCoins(registrationCost),
	); err != nil {
		return nil, errorsmod.Wrap(
			errors.Join(gerrc.ErrInternal, err), "failed to register alias for RollApp",
		)
	}

	return &dymnstypes.MsgRegisterAliasResponse{}, nil
}

// validateRegisterAlias handles validation for the message handled by RegisterAlias.
func (k msgServer) validateRegisterAlias(ctx sdk.Context, msg *dymnstypes.MsgRegisterAlias) error {
	if err := msg.ValidateBasic(); err != nil {
		return err
	}

	rollApp, found := k.rollappKeeper.GetRollapp(ctx, msg.RollappId)
	if !found {
		return errorsmod.Wrapf(gerrc.ErrNotFound, "RollApp: %s", msg.RollappId)
	}

	if rollApp.Owner != msg.Owner {
		return errorsmod.Wrapf(gerrc.ErrPermissionDenied, "not the owner of the RollApp")
	}

	canUseAlias, err := k.CanUseAliasForNewRegistration(ctx, msg.Alias)
	if err != nil {
		return errorsmod.Wrapf(errors.Join(gerrc.ErrInternal, err), "failed to check availability of alias: %s", msg.Alias)
	}

	if !canUseAlias {
		return errorsmod.Wrapf(gerrc.ErrAlreadyExists, "alias already in use or preserved: %s", msg.Alias)
	}

	return nil
}

func (k Keeper) registerAliasForRollApp(
	ctx sdk.Context,
	rollAppId string, owner sdk.AccAddress,
	alias string,
	registrationFee sdk.Coins,
) error {
	if err := k.SetAliasForRollAppId(ctx, rollAppId, alias); err != nil {
		return errorsmod.Wrap(gerrc.ErrInternal, "failed to set alias for RollApp")
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		owner,
		dymnstypes.ModuleName,
		registrationFee,
	); err != nil {
		return err
	}

	if err := k.bankKeeper.BurnCoins(ctx, dymnstypes.ModuleName, registrationFee); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(sdk.NewEvent(
		dymnstypes.EventTypeSell,
		sdk.NewAttribute(dymnstypes.AttributeKeySellAssetType, dymnstypes.TypeAlias.FriendlyString()),
		sdk.NewAttribute(dymnstypes.AttributeKeySellName, alias),
		sdk.NewAttribute(dymnstypes.AttributeKeySellPrice, registrationFee.String()),
		sdk.NewAttribute(dymnstypes.AttributeKeySellTo, rollAppId),
	))

	return nil
}

// EstimateRegisterAlias is a function to estimate the cost of registering an alias.
func EstimateRegisterAlias(
	alias string, params dymnstypes.Params,
) dymnstypes.EstimateRegisterAliasResponse {
	return dymnstypes.EstimateRegisterAliasResponse{
		Price: sdk.NewCoin(params.Price.PriceDenom, params.Price.GetAliasPrice(alias)),
	}
}
