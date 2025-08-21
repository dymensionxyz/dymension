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

// RegisterHLTokenDenomMetadata implements types.MsgServer.
func (m msgServer) RegisterHLTokenDenomMetadata(goCtx context.Context, msg *types.MsgRegisterHLTokenDenomMetadata) (*types.MsgRegisterHLTokenDenomMetadataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	t, err := m.warpK.HypTokens.Get(ctx, msg.HlTokenId.GetInternalId())
	if err != nil {
		return nil, gerrc.ErrNotFound.Wrap("token not found")
	}
	if t.Owner != msg.HlTokenOwner {
		return nil, gerrc.ErrPermissionDenied.Wrap("token owner does not match")
	}
	if t.OriginDenom != msg.TokenMetadata.Base {
		return nil, gerrc.ErrInvalidArgument.Wrap("token origin denom does not match base")
	}
	if err := msg.TokenMetadata.Validate(); err != nil {
		return nil, err
	}

	err = m.CreateDenomMetadata(sdk.UnwrapSDKContext(ctx), msg.TokenMetadata)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterHLTokenDenomMetadataResponse{}, nil
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
