package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

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

	err = uevent.EmitTypedEvent(ctx, &types.EventVote{
		Voter:        msg.Voter,
		Vote:         vote,
		Distribution: distr,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
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

	err = uevent.EmitTypedEvent(ctx, &types.EventRevokeVote{
		Voter:        msg.Voter,
		Distribution: distr,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgRevokeVoteResponse{}, nil
}

func (m MsgServer) UpdateParams(ctx context.Context, msg *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	if msg.Authority != m.k.authority {
		return nil, sdkerrors.ErrorInvalidSigner.Wrapf("Only the gov module can update params")
	}

	oldParams, err := m.k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	err = m.k.SetParams(ctx, msg.NewParams)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	err = uevent.EmitTypedEvent(sdkCtx, &types.EventUpdateParams{
		Authority: msg.Authority,
		NewParams: msg.NewParams,
		OldParams: oldParams,
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
