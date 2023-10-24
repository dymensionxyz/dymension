package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// msgServer provides a way to reference keeper pointer in the message server interface.
type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer for the provided keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

// CreateStream creates a stream and sends coins to the stream.
// Emits create stream event and returns the create stream response.
func (server msgServer) CreateStream(goCtx context.Context, msg *types.MsgCreateStream) (*types.MsgCreateStreamResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	distributeTo, err := sdk.AccAddressFromBech32(msg.DistributeTo)
	if err != nil {
		return nil, err
	}

	streamID, err := server.keeper.CreateStream(ctx, msg.Coins, distributeTo, msg.StartTime, msg.DistrEpochIdentifier, msg.NumEpochsPaidOver)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateStream,
			sdk.NewAttribute(types.AttributeStreamID, osmoutils.Uint64ToString(streamID)),
		),
	})

	return &types.MsgCreateStreamResponse{}, nil
}
