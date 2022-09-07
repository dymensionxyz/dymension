package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, msg.RollappId); isFound {
		return nil, types.ErrRollappExists
	}

	// check to see if there is an active whitelist
	if whitelist := k.DeployerWhitelist(ctx); len(whitelist) > 0 {
		bInWhitelist := false
		// check to see if the creator is in whitelist
		for _, item := range whitelist {
			if item == msg.Creator {
				// Found!
				bInWhitelist = true
				break
			}
		}
		if !bInWhitelist {
			return nil, types.ErrUnauthorizedRollappCreator
		}
	}

	// Create an updated rollapp record
	rollapp := types.Rollapp{
		RollappId:             msg.RollappId,
		Creator:               msg.Creator,
		Version:               0,
		CodeStamp:             msg.CodeStamp,
		GenesisPath:           msg.GenesisPath,
		MaxWithholdingBlocks:  msg.MaxWithholdingBlocks,
		MaxSequencers:         msg.MaxSequencers,
		PermissionedAddresses: msg.PermissionedAddresses,
	}
	// Write rollapp information to the store
	k.SetRollapp(ctx, rollapp)

	return &types.MsgCreateRollappResponse{}, nil
}
