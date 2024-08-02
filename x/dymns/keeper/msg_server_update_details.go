package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "Dym-Name: %s", msg.Name)
	}

	if dymName.IsExpiredAtCtx(ctx) {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "Dym-Name is already expired")
	}

	if dymName.Controller != msg.Controller {
		if dymName.Owner == msg.Controller {
			return nil, errorsmod.Wrapf(
				gerrc.ErrPermissionDenied,
				"please use controller account '%s' to configure", dymName.Controller,
			)
		}

		return nil, gerrc.ErrPermissionDenied
	}

	if msg.Contact == dymnstypes.DoNotModifyDesc && msg.ClearConfigs && len(dymName.Configs) == 0 {
		return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "no existing config to clear")
	}

	return dymName, nil
}
