package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (k msgServer) SetController(goCtx context.Context, msg *dymnstypes.MsgSetController) (*dymnstypes.MsgSetControllerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, err := k.validateSetController(ctx, msg)
	if err != nil {
		return nil, err
	}

	dymName.Controller = msg.Controller
	if err := k.SetDymName(ctx, *dymName); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgSetControllerResponse{}, nil
}

func (k msgServer) validateSetController(ctx sdk.Context, msg *dymnstypes.MsgSetController) (*dymnstypes.DymName, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return nil, dymnstypes.ErrDymNameNotFound.Wrap(msg.Name)
	}

	if dymName.Owner != msg.Owner {
		return nil, sdkerrors.ErrUnauthorized.Wrap("not the owner of the dym name")
	}

	if dymName.IsExpiredAtContext(ctx) {
		return nil, sdkerrors.ErrUnauthorized.Wrap("Dym-Name is already expired")
	}

	if dymName.Controller == msg.Controller {
		return nil, sdkerrors.ErrLogic.Wrap("controller already set")
	}

	return dymName, nil
}
