package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if !k.RollappsEnabled(ctx) {
		return nil, types.ErrRollappsDisabled
	}

	rollappId, err := types.NewChainID(msg.RollappId)
	if err != nil {
		return nil, err
	}

	// error already checked in ValidateBasic
	eip155, _ := types.ParseChainID(msg.RollappId)

	if eip155 == nil {
		return nil, types.ErrInvalidRollappID
	}
	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, rollappId.GetChainID()); isFound {
		return nil, types.ErrRollappExists
	}

	if rollappId.IsEIP155() {
		// check to see if the RollappId has been registered before with same key
		if _, isFound := k.GetRollappByEIP155(ctx, rollappId.GetEIP155ID()); isFound {
			return nil, types.ErrRollappExists
		}
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
