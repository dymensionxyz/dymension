package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// UpdateDetails is message handler,
// handles updating Dym-Name details, performed by the controller.
func (k msgServer) UpdateDetails(goCtx context.Context, msg *dymnstypes.MsgUpdateDetails) (*dymnstypes.MsgUpdateDetailsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, err := k.validateUpdateDetails(ctx, msg)
	if err != nil {
		return nil, err
	}

	var minimumTxGasRequired sdk.Gas

	if msg.Contact == dymnstypes.DoNotModifyDesc {
		minimumTxGasRequired = 0
		// dymName.Contact remaining unchanged
	} else if msg.Contact != "" {
		minimumTxGasRequired = dymnstypes.OpGasUpdateContact
		dymName.Contact = msg.Contact
	} else {
		minimumTxGasRequired = 0
		dymName.Contact = ""
	}

	shouldClearConfigs := msg.ClearConfigs && len(dymName.Configs) > 0

	if shouldClearConfigs {
		dymName.Configs = nil

		if err := k.BeforeDymNameConfigChanged(ctx, dymName.Name); err != nil {
			return nil, err
		}

		if err := k.SetDymName(ctx, *dymName); err != nil {
			return nil, err
		}

		if err := k.AfterDymNameConfigChanged(ctx, dymName.Name); err != nil {
			return nil, err
		}
	} else {
		if err := k.SetDymName(ctx, *dymName); err != nil {
			return nil, err
		}
	}

	consumeMinimumGas(ctx, minimumTxGasRequired, "UpdateDetails")

	return &dymnstypes.MsgUpdateDetailsResponse{}, nil
}

// validateUpdateDetails handles validation for message handled by UpdateDetails
func (k msgServer) validateUpdateDetails(ctx sdk.Context, msg *dymnstypes.MsgUpdateDetails) (*dymnstypes.DymName, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	dymName := k.GetDymName(ctx, msg.Name)
	if dymName == nil {
		return nil, dymnstypes.ErrDymNameNotFound.Wrap(msg.Name)
	}

	if dymName.IsExpiredAtCtx(ctx) {
		return nil, sdkerrors.ErrUnauthorized.Wrap("Dym-Name is already expired")
	}

	if dymName.Controller != msg.Controller {
		if dymName.Owner == msg.Controller {
			return nil, sdkerrors.ErrInvalidAddress.Wrapf(
				"please use controller account '%s' to configure", dymName.Controller,
			)
		}

		return nil, sdkerrors.ErrUnauthorized
	}

	if msg.Contact == dymnstypes.DoNotModifyDesc && msg.ClearConfigs && len(dymName.Configs) == 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrap("no existing config to clear")
	}

	return dymName, nil
}
