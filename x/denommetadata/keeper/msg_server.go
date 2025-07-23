package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	*Keeper
}

// RegisterDenomMetadata implements types.MsgServer.
func (m msgServer) RegisterDenomMetadata(goCtx context.Context, msg *types.MsgRegisterDenomMetadata) (*types.MsgRegisterDenomMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	t, err := m.warpK.HypTokens.Get(ctx, msg.HlTokenId.GetInternalId())
	if err != nil {
		return nil, gerrc.ErrNotFound
	}
	if t.Owner != msg.HlTokenOwner {
		return nil, gerrc.ErrPermissionDenied
	}

	err = m.CreateDenomMetadata(sdk.UnwrapSDKContext(ctx), msg.TokenMetadata)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterDenomMetadataResponse{}, nil
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
