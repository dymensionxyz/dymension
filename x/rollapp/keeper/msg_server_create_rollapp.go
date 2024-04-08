package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.RollappsEnabled(ctx) {
		return nil, types.ErrRollappsDisabled
	}

	err := k.checkIfRollappExists(ctx, msg)
	if err != nil {
		return nil, err
	}

	// check to see if there is an active whitelist
	if whitelist := k.DeployerWhitelist(ctx); len(whitelist) > 0 {
		if !k.IsAddressInDeployerWhiteList(ctx, msg.Creator) {
			return nil, types.ErrUnauthorizedRollappCreator
		}
	}

	rollapp := msg.GetRollapp()
	err = rollapp.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// Write rollapp information to the store
	k.SetRollapp(ctx, rollapp)

	return &types.MsgCreateRollappResponse{}, nil
}

func (k msgServer) checkIfRollappExists(ctx sdk.Context, msg *types.MsgCreateRollapp) error {
	rollappId, err := types.NewChainID(msg.RollappId)
	if err != nil {
		return err
	}
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return types.ErrRollappExists
	}
	if !rollappId.IsEIP155() {
		return nil
	}
	// check to see if the RollappId has been registered before with same key
	rollapp, isFound := k.GetRollappByEIP155(ctx, rollappId.GetEIP155ID())
	// allow replacing EIP155 only when forking (previous rollapp is frozen)
	if !isFound {
		return nil
	}
	if !rollapp.Frozen {
		return types.ErrRollappExists
	}
	previousRollappChainId, _ := types.NewChainID(rollapp.RollappId)
	nextRevisionNumber := previousRollappChainId.GetRevisionNumber() + 1
	if rollappId.GetRevisionNumber() != nextRevisionNumber {
		return errorsmod.Wrapf(types.ErrInvalidRollappID, "revision number should be %d", nextRevisionNumber)
	}
	return nil
}
