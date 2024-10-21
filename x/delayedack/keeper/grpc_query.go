package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

// NewQuerier creates a new Querier struct.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}

// GetPackets implements types.QueryServer.
func (q Querier) GetPackets(goCtx context.Context, req *types.QueryRollappPacketsRequest) (*types.QueryRollappPacketListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	res := &types.QueryRollappPacketListResponse{}

	if req.RollappId == "" {
		// query by status (PENDING by default) and type (if not UNDEFINED)
		res.RollappPackets = q.ListRollappPackets(ctx, types.ByTypeByStatus(req.Type, req.Status))
	} else {
		// query by rollapp id and status (PENDING by default) and type (if not UNDEFINED)
		res.RollappPackets = q.ListRollappPackets(ctx, types.ByRollappIDByTypeByStatus(req.RollappId, req.Type, req.Status))
	}

	// TODO: handle pagination

	return res, nil
}

func (q Querier) GetPendingPacketsByReceiver(goCtx context.Context, req *types.QueryPendingPacketsByReceiverRequest) (*types.QueryPendingPacketByReceiverListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get all pending rollapp packets until the latest finalized height
	rollappPendingPackets, _, err := q.GetPendingPacketsUntilLatestHeight(ctx, req.RollappId)
	if err != nil {
		return nil, fmt.Errorf("get pending rollapp packets until the latest finalized height: rollapp '%s': %w", req.RollappId, err)
	}

	// Filter packets by receiver
	result := make([]commontypes.RollappPacket, 0)
	for _, packet := range rollappPendingPackets {
		// Get packet data
		pd, err := packet.GetTransferPacketData()
		if err != nil {
			return nil, fmt.Errorf("get transfer packet data: rollapp '%s': %w", req.RollappId, err)
		}
		// Return a packet if its receiver matches the one specified
		if pd.Receiver == req.Receiver {
			result = append(result, packet)
		}
	}

	return &types.QueryPendingPacketByReceiverListResponse{
		RollappPackets: result,
		Pagination:     nil, // TODO: handle pagination
	}, nil
}
