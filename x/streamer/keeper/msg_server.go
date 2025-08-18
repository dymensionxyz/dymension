package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

// CreateStream implements the MsgServer interface
func (s msgServer) CreateStream(goCtx context.Context, msg *types.MsgCreateStream) (*types.MsgCreateStreamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", s.authority, msg.Authority)
	}

	streamID, err := s.Keeper.CreateStream(
		ctx,
		msg.Coins,
		msg.DistributeToRecords,
		msg.StartTime,
		msg.DistrEpochIdentifier,
		msg.NumEpochsPaidOver,
		msg.Sponsored,
	)
	if err != nil {
		return nil, err
	}

	if msg.Sponsored && msg.ClearAllVotes {
		err = s.sk.ClearAllVotes(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgCreateStreamResponse{
		StreamId: streamID,
	}, nil
}

// TerminateStream implements the MsgServer interface
func (s msgServer) TerminateStream(goCtx context.Context, msg *types.MsgTerminateStream) (*types.MsgTerminateStreamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", s.authority, msg.Authority)
	}

	err := s.Keeper.TerminateStream(ctx, msg.StreamId)
	if err != nil {
		return nil, err
	}

	return &types.MsgTerminateStreamResponse{}, nil
}

// ReplaceStream implements the MsgServer interface
func (s msgServer) ReplaceStream(goCtx context.Context, msg *types.MsgReplaceStream) (*types.MsgReplaceStreamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", s.authority, msg.Authority)
	}

	err := s.ReplaceDistrRecords(ctx, msg.StreamId, msg.Records)
	if err != nil {
		return nil, err
	}

	return &types.MsgReplaceStreamResponse{}, nil
}

// UpdateStream implements the MsgServer interface
func (s msgServer) UpdateStream(goCtx context.Context, msg *types.MsgUpdateStream) (*types.MsgUpdateStreamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", s.authority, msg.Authority)
	}

	err := s.UpdateDistrRecords(ctx, msg.StreamId, msg.Records)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateStreamResponse{}, nil
}

// UpdateParams is a governance operation to update the module parameters.
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the sender is the authority
	if req.Authority != m.authority {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only the gov module can update params")
	}

	err := req.Params.ValidateBasic()
	if err != nil {
		return nil, err
	}

	m.SetParams(ctx, req.Params)
	return &types.MsgUpdateParamsResponse{}, nil
}
