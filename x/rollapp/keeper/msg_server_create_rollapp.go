package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

func (k msgServer) CreateRollapp(goCtx context.Context, msg *types.MsgCreateRollapp) (*types.MsgCreateRollappResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// get the creator address
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	// check to see if the RollappId has been registered before
	if _, isFound := k.GetRollapp(ctx, msg.RollappId); isFound {
		return nil, types.ErrRollappExists
	}

	// check sequencers addresses
	for _, addr := range msg.PermissionedAddresses.Addresses {
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cosmos.AddressString, got %T", addr)
		}
	}

	// Create an updated whois record
	rollapp := types.Rollapp{
		RollappId:             msg.RollappId,
		Creator:               creator.String(),
		Version:               0,
		CodeStamp:             msg.CodeStamp,
		GenesisPath:           msg.GenesisPath,
		MaxWithholdingBlocks:  msg.MaxWithholdingBlocks,
		MaxSequencers:         msg.MaxSequencers,
		PermissionedAddresses: msg.PermissionedAddresses,
	}
	// Write whois information to the store
	k.SetRollapp(ctx, rollapp)

	return &types.MsgCreateRollappResponse{}, nil
}
