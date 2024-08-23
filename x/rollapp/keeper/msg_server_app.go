package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// AddApp adds a new app
func (k msgServer) AddApp(goCtx context.Context, msg *types.MsgAddApp) (*types.MsgAddAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.checkInputs(ctx, msg); err != nil {
		return nil, err
	}

	// signal to client that ordering is up to them
	if msg.Order == 0 {
		msg.Order = -1
	}

	// charge the app creation fee
	creator := sdk.MustAccAddressFromBech32(msg.Creator)
	appCost := sdk.NewCoins(k.AppCreationCost(ctx))

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, appCost); err != nil {
		return nil, types.ErrAppCreationCostPayment
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, appCost); err != nil {
		return nil, types.ErrAppCreationCostPayment
	}

	app := msg.GetApp()
	k.SetApp(ctx, app)

	if err := ctx.EventManager().EmitTypedEvent(app.GetAddedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgAddAppResponse{}, nil
}

// UpdateApp updates an existing app
func (k msgServer) UpdateApp(goCtx context.Context, msg *types.MsgUpdateApp) (*types.MsgUpdateAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.checkInputs(ctx, msg); err != nil {
		return nil, err
	}

	app := msg.GetApp()

	k.SetApp(ctx, app)

	if err := ctx.EventManager().EmitTypedEvent(app.GetUpdatedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgUpdateAppResponse{}, nil
}

// RemoveApp deletes an existing app
func (k msgServer) RemoveApp(goCtx context.Context, msg *types.MsgRemoveApp) (*types.MsgRemoveAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.checkInputs(ctx, msg); err != nil {
		return nil, err
	}

	app := msg.GetApp()

	k.DeleteApp(ctx, app)

	if err := ctx.EventManager().EmitTypedEvent(app.GetRemovedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgRemoveAppResponse{}, nil
}

func (k msgServer) checkInputs(ctx sdk.Context, msg appMsg) error {
	rollapp, foundRollapp := k.GetRollapp(ctx, msg.GetRollappId())
	if !foundRollapp {
		return errorsmod.Wrapf(types.ErrNotFound, "rollappId=%s", msg.GetRollappId())
	}

	// check if the sender is the owner of the app
	if msg.GetCreator() != rollapp.Owner {
		return types.ErrUnauthorizedSigner
	}

	// check if the app already exists
	if _, foundApp := k.GetApp(ctx, msg.GetName(), msg.GetRollappId()); foundApp {
		return types.ErrAppExists
	}

	return nil
}

type appMsg interface {
	GetName() string
	GetRollappId() string
	GetCreator() string
}

var _ types.MsgServer = msgServer{}
