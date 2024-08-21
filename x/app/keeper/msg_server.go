package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/app/types"
)

type msgServer struct {
	Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return &msgServer{Keeper: k}
}

// CreateApp creates a new app
func (m msgServer) CreateApp(goCtx context.Context, msg *types.MsgCreateApp) (*types.MsgCreateAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := m.checkInputsAndGetApp(ctx, msg)
	if err != nil {
		return nil, types.ErrNotFound
	}

	// charge the app creation fee
	creator := sdk.MustAccAddressFromBech32(msg.Creator)
	regFee := sdk.NewCoins(m.Cost(ctx))

	if err = m.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, regFee); err != nil {
		return nil, types.ErrFeePayment
	}

	if err = m.bankKeeper.BurnCoins(ctx, types.ModuleName, regFee); err != nil {
		return nil, types.ErrFeePayment
	}

	app := msg.GetApp()
	m.SetApp(ctx, app)

	if err = ctx.EventManager().EmitTypedEvent(app.GetCreatedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgCreateAppResponse{}, nil
}

// UpdateApp updates an existing app
func (m msgServer) UpdateApp(goCtx context.Context, msg *types.MsgUpdateApp) (*types.MsgUpdateAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := m.checkInputsAndGetApp(ctx, msg)
	if err != nil {
		return nil, types.ErrNotFound
	}

	app := msg.GetApp()

	m.SetApp(ctx, app)

	if err = ctx.EventManager().EmitTypedEvent(app.GetUpdatedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgUpdateAppResponse{}, nil
}

// DeleteApp deletes an existing app
func (m msgServer) DeleteApp(goCtx context.Context, msg *types.MsgDeleteApp) (*types.MsgDeleteAppResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := m.checkInputsAndGetApp(ctx, msg)
	if err != nil {
		return nil, types.ErrNotFound
	}

	app := msg.GetApp()

	m.RemoveApp(ctx, app)

	if err = ctx.EventManager().EmitTypedEvent(app.GetDeletedEvent()); err != nil {
		return nil, err
	}

	return &types.MsgDeleteAppResponse{}, nil
}

type appMsg interface {
	GetName() string
	GetRollappId() string
	GetCreator() string
}

func (m msgServer) checkInputsAndGetApp(ctx sdk.Context, msg appMsg) (*types.App, error) {
	rollapp, ok := m.rollappKeeper.GetRollapp(ctx, msg.GetRollappId())
	if !ok {
		return nil, types.ErrRollappNotFound
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

var _ types.MsgServer = msgServer{}
