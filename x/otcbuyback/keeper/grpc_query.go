package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

var _ types.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) types.QueryServer {
	return &queryServer{Keeper: keeper}
}

// AllAuctions queries all auctions with optional filtering
func (q queryServer) AllAuctions(goCtx context.Context, req *types.QueryAllAuctionsRequest) (*types.QueryAllAuctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	auctions, err := q.GetAllAuctions(ctx, req.ExcludeCompleted)
	if err != nil {
		return nil, err
	}

	return &types.QueryAllAuctionsResponse{
		Auctions: auctions,
	}, nil
}

// Auction queries a specific auction by ID
func (q queryServer) Auction(goCtx context.Context, req *types.QueryAuctionRequest) (*types.QueryAuctionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	auction, found := q.GetAuction(ctx, req.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "auction with id %d not found", req.Id)
	}

	// Calculate current discount percentage
	currentDiscount := auction.GetCurrentDiscount(ctx.BlockTime())

	return &types.QueryAuctionResponse{
		Auction:         auction,
		CurrentDiscount: currentDiscount,
	}, nil
}

// UserPurchase queries user's purchase for a specific auction
func (q queryServer) UserPurchase(goCtx context.Context, req *types.QueryUserPurchaseRequest) (*types.QueryUserPurchaseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	user, err := sdk.AccAddressFromBech32(req.User)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user address: %s", req.User)
	}

	auction, found := q.GetAuction(ctx, req.AuctionId)
	if !found {
		return nil, status.Errorf(codes.NotFound, "auction with id %d not found", req.AuctionId)
	}

	purchase, found := q.GetPurchase(ctx, req.AuctionId, user)
	if !found {
		return nil, status.Errorf(codes.NotFound, "no purchase found for user %s in auction %d", req.User, req.AuctionId)
	}

	// Calculate claimable amount
	claimableAmount := purchase.VestedAmount(ctx.BlockTime(), auction.GetVestingStartTime(), auction.GetVestingEndTime())

	return &types.QueryUserPurchaseResponse{
		Purchase:        purchase,
		ClaimableAmount: claimableAmount,
	}, nil
}

// AcceptedTokens queries all accepted tokens with their current prices
func (q queryServer) AcceptedTokens(goCtx context.Context, req *types.QueryAcceptedTokensRequest) (*types.QueryAcceptedTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	acceptedTokens, err := q.GetAllAcceptedTokens(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryAcceptedTokensResponse{
		AcceptedTokens: acceptedTokens,
	}, nil
}

// AcceptedToken queries specific token data and price by denom
func (q queryServer) AcceptedToken(goCtx context.Context, req *types.QueryAcceptedTokenRequest) (*types.QueryAcceptedTokenResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tokenData, err := q.GetAcceptedTokenData(ctx, req.Denom)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "token %s not found in accepted tokens", req.Denom)
	}

	acceptedToken := types.AcceptedToken{
		Denom:     req.Denom,
		TokenData: tokenData,
	}
	// Get current spot price (AMM)
	spotPrice, err := q.ammKeeper.CalculateSpotPrice(ctx, tokenData.PoolId, req.Denom, q.baseDenom)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get spot price for token %s: %v", req.Denom, err)
	}

	return &types.QueryAcceptedTokenResponse{
		AcceptedToken: acceptedToken,
		SpotPrice:     spotPrice,
	}, nil
}
