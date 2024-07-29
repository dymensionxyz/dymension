package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

var _ types.MsgServer = MsgServer{}

type MsgServer struct {
	k Keeper
}

func NewMsgServer(k Keeper) MsgServer {
	return MsgServer{k: k}
}

func (m MsgServer) Vote(goCtx context.Context, msg *types.MsgVote) (*types.MsgVoteResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// Don't check the error since it's part of validation
	voter := sdk.MustAccAddressFromBech32(msg.Voter)

	vote, distr, err := m.k.Vote(ctx, voter, msg.Weights)
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventVote{
		Voter:        msg.Voter,
		Vote:         vote,
		Distribution: distr,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgVoteResponse{}, nil
}

func (m MsgServer) RevokeVote(goCtx context.Context, msg *types.MsgRevokeVote) (*types.MsgRevokeVoteResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	// Don't check the error since it's part of validation
	voter := sdk.MustAccAddressFromBech32(msg.Voter)

	distr, err := m.k.RevokeVote(ctx, voter)
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventRevokeVote{
		Voter:        msg.Voter,
		Distribution: distr,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRevokeVoteResponse{}, nil
}
