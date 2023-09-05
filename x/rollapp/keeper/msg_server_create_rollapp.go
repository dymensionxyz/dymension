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
		var item types.DeployerParams
		for _, item = range whitelist {
			if item.Address == msg.Creator {
				// Found!
				bInWhitelist = true
				break
			}
		}
		if !bInWhitelist {
			return nil, types.ErrUnauthorizedRollappCreator
		}

		if item.MaxRollapps > 0 {
			// if MaxRollapps, it means there is a limit for this creator
			// count how many rollapps he created
			rollappsNumOfCreator := uint64(0)
			for _, r := range k.GetAllRollapp(ctx) {
				if r.Creator == msg.Creator {
					rollappsNumOfCreator += 1
				}
			}
			// check the creator didn't hit the maximum
			if rollappsNumOfCreator >= item.MaxRollapps {
				// check the deployer max rollapps limitation
				return nil, types.ErrRollappCreatorExceedMaximumRollapps
			}
		}
	}

	// Create an updated rollapp record
	rollapp := types.Rollapp{
		RollappId:             msg.RollappId,
		Creator:               msg.Creator,
		Version:               0,
		MaxSequencers:         msg.MaxSequencers,
		PermissionedAddresses: msg.PermissionedAddresses,
	}
	// Write rollapp information to the store
	k.SetRollapp(ctx, rollapp)

	return &types.MsgCreateRollappResponse{}, nil
}
