package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func (k msgServer) UpdateResolveAddress(goCtx context.Context, msg *dymnstypes.MsgUpdateResolveAddress) (*dymnstypes.MsgUpdateResolveAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName, err := k.validateUpdateResolveAddress(ctx, msg)
	if err != nil {
		return nil, err
	}

	_, newConfig := msg.GetDymNameConfig()
	if newConfig.ChainId == ctx.ChainID() {
		newConfig.ChainId = ""
	}
	newConfigIdentity := newConfig.GetIdentity()

	var minimumTxGasRequired sdk.Gas

	existingConfigCount := len(dymName.Configs)
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
			// no-config case also falls into this branch

			// do nothing
		} else {
			if existingConfigCount == 1 {
				dymName.Configs = nil
			} else {
				dymName.Configs[foundSameConfigIdAtIdx] = dymName.Configs[existingConfigCount-1]
				dymName.Configs = dymName.Configs[:existingConfigCount-1]
			}
		}
	} else {
		minimumTxGasRequired = dymnstypes.OpGasConfig

		if existingConfigCount > 0 {
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
		} else {
			dymName.Configs = []dymnstypes.DymNameConfig{newConfig}
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

	consumeMinimumGas(ctx, minimumTxGasRequired, "UpdateResolveAddress")

	return &dymnstypes.MsgUpdateResolveAddressResponse{}, nil
}

func (k msgServer) validateUpdateResolveAddress(ctx sdk.Context, msg *dymnstypes.MsgUpdateResolveAddress) (*dymnstypes.DymName, error) {
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

	if msg.ResolveTo != "" && (msg.ChainId == "" || msg.ChainId == ctx.ChainID()) {
		if _, err := sdk.AccAddressFromBech32(msg.ResolveTo); err != nil {
			return nil, sdkerrors.ErrInvalidAddress.Wrap(
				"resolve address must be a valid bech32 account address on host chain",
			)
		}
	}

	return dymName, nil
}
