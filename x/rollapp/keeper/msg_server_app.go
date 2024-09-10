package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// AddApp adds a new app
func (k msgServer) AddApp(goCtx context.Context, msg *types.MsgAddApp) (*types.MsgAddAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.checkInputs(ctx, msg); err != nil {
		return nil, err
	}

	// If order is not set by the client, all order will get -1 which will make it random.
	if msg.Order == 0 {
		msg.Order = -1
	}

	// charge the app registration fee
	creator := sdk.MustAccAddressFromBech32(msg.Creator)
	appFee := sdk.NewCoins(k.AppRegistrationFee(ctx))

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, appFee); err != nil {
		return nil, types.ErrAppRegistrationFeePayment
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, appFee); err != nil {
		return nil, types.ErrAppRegistrationFeePayment
	}

	app := msg.GetApp()
	app.Id = k.GetNextAppID(ctx, app.RollappId)

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
	app := msg.GetApp()
	rollapp, foundRollapp := k.GetRollapp(ctx, app.GetRollappId())
	if !foundRollapp {
		return gerrc.ErrNotFound.Wrapf("rollappId: %s", app.GetRollappId())
	}

	// check if the sender is the owner of the app
	if msg.GetCreator() != rollapp.Owner {
		return gerrc.ErrPermissionDenied.Wrap("not the owner of the RollApp")
	}

	switch msg.(type) {
	case *types.MsgRemoveApp, *types.MsgUpdateApp:
		if idExists := k.appIDExists(ctx, app); !idExists {
			return gerrc.ErrNotFound.Wrap("app not found")
		}
	}
	switch msg.(type) {
	case *types.MsgAddApp, *types.MsgUpdateApp:
		apps := k.GetRollappApps(ctx, app.GetRollappId())
		if nameExists := k.appNameExists(apps, app); nameExists {
			return gerrc.ErrAlreadyExists.Wrap("app name already exists")
		}
	}

	return nil
}

func (k msgServer) appNameExists(apps []*types.App, app types.App) bool {
	for _, a := range apps {
		// does name already exist:
		// - id=0 means it is a new app
		// - skip if the id is the same as the app being checked
		if (app.GetId() == 0 || a.Id != app.GetId()) && a.Name == app.GetName() {
			return true
		}
	}
	return false
}

func (k msgServer) appIDExists(ctx sdk.Context, app types.App) bool {
	_, foundApp := k.GetApp(ctx, app.GetId(), app.GetRollappId())
	return foundApp
}

type appMsg interface {
	GetCreator() string
	GetApp() types.App
}

var _ types.MsgServer = msgServer{}
