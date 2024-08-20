package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// SetController is message handler,
// handles setting a controller for a Dym-Name, performed by the owner.
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

// validateSetController handles validation for message handled by SetController
func (k msgServer) validateSetController(ctx sdk.Context, msg *dymnstypes.MsgSetController) (*dymnstypes.DymName, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.Name)
	}

	if dymName.Owner != msg.Owner {
		return nil, errorsmod.Wrap(gerrc.ErrPermissionDenied, "not the owner of the Dym-Name")
	}

	if dymName.IsExpiredAtCtx(ctx) {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "Dym-Name is already expired")
	}

	if dymName.Controller == msg.Controller {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "controller already set")
	}

	return dymName, nil
}
