package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/kas/types"

	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type msgServer struct {
	*Keeper
}

func (m msgServer) Foo(context.Context, *types.MsgFoo) (*types.MsgFooResponse, error) {
	panic("unimplemented")
}

func (k *Keeper) IndicateProgress(goCtx context.Context, req *types.MsgIndicateProgress) (*types.MsgIndicateProgressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	metadata, err := hypercoretypes.NewMessageIdMultisigRawMetadata(req.Metadata)
	if err != nil {
		return nil, gerrc.ErrInvalidArgument.Wrapf("metadata: %w", err)
	}

	if req.GetPayload() == nil {
		return nil, gerrc.ErrInvalidArgument.Wrapf("payload is nil")
	}

	digest, err := req.GetPayload().SignBytes()
	if err != nil {
		return nil, gerrc.ErrInvalidArgument.Wrapf("payload: %w", err)
	}
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
