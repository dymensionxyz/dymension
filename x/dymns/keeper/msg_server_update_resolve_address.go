package keeper

import (
	"context"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// UpdateResolveAddress is message handler,
// handles updating Dym-Name-Address resolution configuration, performed by the controller.
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

	if newConfig.ChainId == "" || k.IsRollAppId(ctx, newConfig.ChainId) {
		// guarantee of case-insensitive on host and RollApps,
		// so we do normalize input
		newConfig.Value = strings.ToLower(newConfig.Value)
	} else if dymnsutils.IsValidHexAddress(newConfig.Value) {
		// if the address is hex format, then treat the chain is case-insensitive address,
		// like Ethereum, where the address is case-insensitive and checksum address contains mixed case
		newConfig.Value = strings.ToLower(newConfig.Value)
	}

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
			dymName.Configs = append(
				dymName.Configs[:foundSameConfigIdAtIdx],
				dymName.Configs[foundSameConfigIdAtIdx+1:]...,
			)
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

	// Charge protocol fee.
	// The protocol fee mechanism is used to prevent spamming to the network.
	consumeMinimumGas(ctx, minimumTxGasRequired, "UpdateResolveAddress")

	return &dymnstypes.MsgUpdateResolveAddressResponse{}, nil
}

// validateUpdateResolveAddress handles validation for message handled by UpdateResolveAddress
func (k msgServer) validateUpdateResolveAddress(ctx sdk.Context, msg *dymnstypes.MsgUpdateResolveAddress) (*dymnstypes.DymName, error) {
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

	if msg.ResolveTo != "" {
		if msg.ChainId == "" || msg.ChainId == ctx.ChainID() {
			if !dymnsutils.IsValidBech32AccountAddress(msg.ResolveTo, true) {
				return nil, errorsmod.Wrap(
					gerrc.ErrInvalidArgument,
					"resolve address must be a valid bech32 account address on host chain",
				)
			}
		} else if k.IsRollAppId(ctx, msg.ChainId) {
			if !dymnsutils.IsValidBech32AccountAddress(msg.ResolveTo, false) {
				return nil, errorsmod.Wrap(
					gerrc.ErrInvalidArgument,
					"resolve address must be a valid bech32 account address on RollApp",
				)
			}
			if bech32Prefix, found := k.GetRollAppBech32Prefix(ctx, msg.ChainId); found {
				hrp, _, err := bech32.DecodeAndConvert(msg.ResolveTo)
				if err != nil {
					panic("unreachable")
				}
				if hrp != bech32Prefix {
					return nil, errorsmod.Wrapf(
						gerrc.ErrInvalidArgument,
						"resolve address must be a valid bech32 account address on RollApps: %s", bech32Prefix,
					)
				}
			}
		}
	}

	return dymName, nil
}
