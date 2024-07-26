package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k msgServer) UpdateRollappInformation(goCtx context.Context, msg *types.MsgUpdateRollappInformation) (*types.MsgUpdateRollappInformationResponse, error) {
	if msg == nil {
		return nil, types.ErrInvalidRequest
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := k.UpdateRollapp(ctx, msg); err != nil {
		return nil, err
	}

	return &types.MsgUpdateRollappInformationResponse{}, nil
}
