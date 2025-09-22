package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/math"
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

var _ types.QueryServer = queryServer{}

// FeeHook returns a fee hook by ID
func (k queryServer) FeeHook(ctx context.Context, req *types.QueryFeeHookRequest) (*types.QueryFeeHookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	hookId, err := util.DecodeHexAddress(req.Id)
	if err != nil {
		return nil, err
	}

	hook, err := k.feeHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, fmt.Errorf("get fee hook: %w", err)
	}

	return &types.QueryFeeHookResponse{FeeHook: hook}, nil
}

// FeeHooks returns all fee hooks
func (k queryServer) FeeHooks(ctx context.Context, req *types.QueryFeeHooksRequest) (*types.QueryFeeHooksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	values, pagination, err := util.GetPaginatedFromMap(ctx, k.feeHooks, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryFeeHooksResponse{
		FeeHooks:   values,
		Pagination: pagination,
	}, nil
}

// AggregationHook returns an aggregation hook by ID
func (k queryServer) AggregationHook(ctx context.Context, req *types.QueryAggregationHookRequest) (*types.QueryAggregationHookResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	hookId, err := util.DecodeHexAddress(req.Id)
	if err != nil {
		return nil, err
	}

	hook, err := k.aggregationHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, fmt.Errorf("get aggregation hook: %w", err)
	}

	return &types.QueryAggregationHookResponse{AggregationHook: hook}, nil
}

// AggregationHooks returns all aggregation hooks
func (k queryServer) AggregationHooks(ctx context.Context, req *types.QueryAggregationHooksRequest) (*types.QueryAggregationHooksResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	values, pagination, err := util.GetPaginatedFromMap(ctx, k.aggregationHooks, req.Pagination)
	if err != nil {
		return nil, err
	}

	return &types.QueryAggregationHooksResponse{
		AggregationHooks: values,
		Pagination:       pagination,
	}, nil
}

// QuoteFeePayment quotes the fee payment required for a transfer
func (k queryServer) QuoteFeePayment(ctx context.Context, req *types.QueryQuoteFeePaymentRequest) (*types.QueryQuoteFeePaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	hookId, err := util.DecodeHexAddress(req.HookId)
	if err != nil {
		return nil, fmt.Errorf("decode hook_id: %w", err)
	}

	tokenId, err := util.DecodeHexAddress(req.TokenId)
	if err != nil {
		return nil, fmt.Errorf("decode token_id: %w", err)
	}

	transferAmt, ok := math.NewIntFromString(req.TransferAmount)
	if !ok {
		return nil, errors.New("failed to convert transfer_amount to math.Int")
	}

	feeHandler := NewFeeHookHandler(k.Keeper)

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	fee, err := feeHandler.QuoteFee(sdkCtx, hookId, tokenId, transferAmt)
	if err != nil {
		return nil, fmt.Errorf("quote fee in base: %w", err)
	}

	return &types.QueryQuoteFeePaymentResponse{
		FeeCoins: fee,
	}, nil
}
