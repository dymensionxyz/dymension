package keeper

import (
	"context"

    "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) SubmitFraud(goCtx context.Context,  msg *types.MsgSubmitFraud) (*types.MsgSubmitFraudResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgSubmitFraudResponse{}, nil
}
