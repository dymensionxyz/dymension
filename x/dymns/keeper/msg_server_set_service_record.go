package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// SetServiceRecord is message handler,
// handles setting a typed service/endpoint record on a Dym-Name, performed by the controller.
func (k msgServer) SetServiceRecord(goCtx context.Context, msg *dymnstypes.MsgSetServiceRecord) (*dymnstypes.MsgSetServiceRecordResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	originalConsumedGas := ctx.GasMeter().GasConsumed()

	dymName, err := k.validateSetServiceRecord(ctx, msg)
	if err != nil {
		return nil, err
	}

	_, newConfig := msg.GetDymNameConfig()
	newConfigIdentity := newConfig.GetIdentity()

	var minimumTxGasRequired storetypes.Gas

	if newConfig.IsDelete() {
		minimumTxGasRequired = 0 // do not charge for delete

		foundSameConfigIdAtIdx := -1
		for i, config := range dymName.Configs {
			if config.GetIdentity() == newConfigIdentity {
				foundSameConfigIdAtIdx = i
				break
			}
		}

		if foundSameConfigIdAtIdx < 0 {
			return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "config")
		}

		dymName.Configs = append(
			dymName.Configs[:foundSameConfigIdAtIdx],
			dymName.Configs[foundSameConfigIdAtIdx+1:]...,
		)
	} else {
		minimumTxGasRequired = dymnstypes.OpGasConfig

		var foundSameConfigId bool
		for i, config := range dymName.Configs {
			if config.GetIdentity() == newConfigIdentity {
				dymName.Configs[i] = newConfig
				foundSameConfigId = true
				break
			}
		}
		if !foundSameConfigId {
			dymName.Configs = append(dymName.Configs, newConfig)
		}
	}

	if err := k.BeforeDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return nil, err
	}

	if err := k.SetDymName(ctx, *dymName); err != nil {
		return nil, err
	}

	if err := k.AfterDymNameConfigChanged(ctx, dymName.Name); err != nil {
		return nil, err
	}

	// Charge protocol fee.
	// The protocol fee mechanism is used to prevent spamming to the network.
	consumeMinimumGas(ctx, minimumTxGasRequired, originalConsumedGas, "SetServiceRecord")

	return &dymnstypes.MsgSetServiceRecordResponse{}, nil
}

// validateSetServiceRecord handles validation for message handled by SetServiceRecord
func (k msgServer) validateSetServiceRecord(ctx sdk.Context, msg *dymnstypes.MsgSetServiceRecord) (*dymnstypes.DymName, error) {
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
			return nil, errorsmod.Wrapf(gerrc.ErrPermissionDenied,
				"please use controller account '%s' to configure", dymName.Controller,
			)
		}

		return nil, gerrc.ErrPermissionDenied
	}

	return dymName, nil
}
