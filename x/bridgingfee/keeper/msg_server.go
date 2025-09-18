package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// CreateBridgingFeeHook creates a new bridging fee hook
func (k msgServer) CreateBridgingFeeHook(goCtx context.Context, msg *types.MsgCreateBridgingFeeHook) (*types.MsgCreateBridgingFeeHookResponse, error) {
	hookId, err := k.CreateFeeHook(goCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateBridgingFeeHookResponse{Id: hookId}, nil
}

// SetBridgingFeeHook updates an existing bridging fee hook
func (k msgServer) SetBridgingFeeHook(goCtx context.Context, msg *types.MsgSetBridgingFeeHook) (*types.MsgSetBridgingFeeHookResponse, error) {
	err := k.UpdateFeeHook(goCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetBridgingFeeHookResponse{}, nil
}

// CreateAggregationHook creates a new aggregation hook
func (k msgServer) CreateAggregationHook(goCtx context.Context, msg *types.MsgCreateAggregationHook) (*types.MsgCreateAggregationHookResponse, error) {
	hookId, err := k.Keeper.CreateAggregationHook(goCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateAggregationHookResponse{Id: hookId}, nil
}

// SetAggregationHook updates an existing aggregation hook
func (k msgServer) SetAggregationHook(goCtx context.Context, msg *types.MsgSetAggregationHook) (*types.MsgSetAggregationHookResponse, error) {
	err := k.UpdateAggregationHook(goCtx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgSetAggregationHookResponse{}, nil
}
