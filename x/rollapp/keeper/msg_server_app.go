package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// AddApp adds a new app
func (m msgServer) AddApp(goCtx context.Context, msg *types.MsgAddApp) (*types.MsgAddAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := m.checkInputsAndGetApp(ctx, msg)
	if err != nil {
		return nil, err
	}

	// charge the app creation fee
	creator := sdk.MustAccAddressFromBech32(msg.Creator)
	regFee := sdk.NewCoins(m.AppCost(ctx))

	if err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, regFee); err != nil {
		return nil, types.ErrAppCostPayment
	}

	if err = m.bankKeeper.BurnCoins(ctx, types.ModuleName, regFee); err != nil {
		return nil, types.ErrAppCostPayment
	}

	app := msg.GetApp()
	m.SetApp(ctx, app)

	if err = ctx.EventManager().EmitTypedEvent(app.GetAddedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgAddAppResponse{}, nil
}

// UpdateApp updates an existing app
func (m msgServer) UpdateApp(goCtx context.Context, msg *types.MsgUpdateApp) (*types.MsgUpdateAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := m.checkInputsAndGetApp(ctx, msg)
	if err != nil {
		return nil, err
	}

	app := msg.GetApp()

	m.SetApp(ctx, app)

	if err = ctx.EventManager().EmitTypedEvent(app.GetUpdatedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgUpdateAppResponse{}, nil
}

// RemoveApp deletes an existing app
func (m msgServer) RemoveApp(goCtx context.Context, msg *types.MsgRemoveApp) (*types.MsgRemoveAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := m.checkInputsAndGetApp(ctx, msg)
	if err != nil {
		return nil, err
	}

	app := msg.GetApp()

	m.DeleteApp(ctx, app)

	if err = ctx.EventManager().EmitTypedEvent(app.GetRemovedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgRemoveAppResponse{}, nil
}

func (m msgServer) checkInputsAndGetApp(ctx sdk.Context, msg appMsg) (*types.App, error) {
	rollapp, ok := m.GetRollapp(ctx, msg.GetRollappId())
	if !ok {
		return nil, errorsmod.Wrapf(types.ErrNotFound, "rollappId=%s", msg.GetRollappId())
	}

	// check if the sender is the owner of the app
	if msg.GetCreator() != rollapp.Owner {
		return nil, types.ErrUnauthorizedSigner
	}

	// check if the app already exists
	app, ok := m.GetApp(ctx, msg.GetName(), msg.GetRollappId())
	if ok {
		return nil, types.ErrAppExists
	}

	return &app, nil
}

type appMsg interface {
	GetName() string
	GetRollappId() string
	GetCreator() string
}

var _ types.MsgServer = msgServer{}
