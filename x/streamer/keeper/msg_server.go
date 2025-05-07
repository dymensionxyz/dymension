package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	k Keeper
}

func NewMsgServerImpl(k Keeper) types.MsgServer {
	return MsgServer{k: k}
}

// UpdateParams is a governance operation to update the module parameters.
func (m MsgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the sender is the authority
	if req.Authority != m.k.authority {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only the gov module can update params")
	}

	err := req.Params.ValidateBasic()
	if err != nil {
		return nil, err
	}

	m.k.SetParams(ctx, req.Params)
	return &types.MsgUpdateParamsResponse{}, nil
}
