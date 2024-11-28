package keeper

import (
	"context"
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// UpdateRollappInformation updates the rollapp information
// It allows to change:
// - the rollapp metadata
// - the genesis info (in case the genesis info is not sealed)
// - the initial sequencer (in case the rollapp is not launched)
func (k msgServer) UpdateRollappInformation(goCtx context.Context, msg *types.MsgUpdateRollappInformation) (*types.MsgUpdateRollappInformationResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	updated, err := k.CheckAndUpdateRollappFields(ctx, msg)
	if err != nil {
		return nil, err
	}

	k.SetRollapp(ctx, updated)

	if err = uevent.EmitTypedEvent(ctx, msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateRollappInformationResponse{}, nil
}

// ForceGenesisInfoChange allows the gov module to force a genesis info change, even if the genesis info is already sealed
func (k Keeper) ForceGenesisInfoChange(goCtx context.Context, msg *types.MsgForceGenesisInfoChange) (*types.MsgForceGenesisInfoChangeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != k.authority {
		err := errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can submit proposals")
		ctx.Logger().Error("force genesis info change.", err)
		return nil, err
	}

	// validateBasic check that the new genesis info is valid and contains all the necessary fields
	if err := msg.ValidateBasic(); err != nil {
		err = errors.Join(gerrc.ErrInvalidArgument, err)
		ctx.Logger().Error("force genesis info change.", err)
		return nil, err
	}

	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		err := errorsmod.Wrapf(types.ErrRollappNotFound, "rollapp not found: %s", msg.RollappId)
		ctx.Logger().Error("force genesis info change.", err)
		return nil, err
	}

	rollapp.GenesisInfo = msg.NewGenesisInfo
	rollapp.GenesisInfo.Sealed = true
	k.SetRollapp(ctx, rollapp)

	if err := uevent.EmitTypedEvent(ctx, msg); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgForceGenesisInfoChangeResponse{}, nil
}
