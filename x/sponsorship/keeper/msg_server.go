package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	k Keeper
}

func NewMsgServer(k Keeper) MsgServer {
	return MsgServer{k: k}
}

func (m MsgServer) UpdateParams(ctx context.Context, params *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (m MsgServer) Vote(ctx context.Context, vote *types.MsgVote) (*types.MsgVoteResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (m MsgServer) RevokeVote(ctx context.Context, vote *types.MsgRevokeVote) (*types.MsgRevokeVoteResponse, error) {
	// TODO implement me
	panic("implement me")
}
