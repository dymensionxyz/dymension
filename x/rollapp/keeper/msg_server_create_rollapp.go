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

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, msg.RollappId); isFound {
		return nil, types.ErrRollappExists
	}

	// check to see if there is an active whitelist
	if whitelist := k.DeployerWhitelist(ctx); len(whitelist) > 0 {
		if !k.IsAddressInDeployerWhiteList(ctx, msg.Creator) {
			return nil, types.ErrUnauthorizedRollappCreator
		}
	}

	// Build the genesis state from the genesis accounts
	var rollappGenesisState *types.RollappGenesisState
	if len(msg.GenesisAccounts) > 0 {
		rollappGenesisState = &types.RollappGenesisState{
			GenesisAccounts: msg.GenesisAccounts,
			IsGenesisEvent:  false,
		}
	}

	// copy TokenMetadata
	metadata := make([]*types.TokenMetadata, len(msg.Metadatas))
	for i := range msg.Metadatas {
		metadata[i] = &msg.Metadatas[i]
	}

	// Create an updated rollapp record
	rollapp := types.NewRollapp(msg.Creator, msg.RollappId, msg.MaxSequencers, msg.PermissionedAddresses, metadata, rollappGenesisState)

	// Write rollapp information to the store
	k.SetRollapp(ctx, rollapp)

	return &types.MsgCreateRollappResponse{}, nil
}
