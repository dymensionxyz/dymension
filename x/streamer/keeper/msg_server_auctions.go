package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	hatypes "github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// CreateAuction implements the MsgServer interface for creating Dutch auctions
func (s msgServer) CreateAuction(goCtx context.Context, msg *types.MsgCreateAuction) (*types.MsgCreateAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", s.authority, msg.Authority)
	}

	// FIXME: move funds from the module account to the auction account

	// Create the auction and stream
	auctionID, err := s.Keeper.ahk.CreateAuction(
		ctx,
		msg.Allocation,
		msg.StartTime,
		msg.EndTime,
		msg.InitialDiscount,
		msg.MaxDiscount,
		hatypes.Auction_VestingPlan{
			VestingPeriod:               msg.VestingPeriod,
			VestingStartAfterAuctionEnd: msg.VestingStartAfterAuctionEnd,
		},
	)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateAuctionResponse{
		AuctionId: auctionID,
	}, nil
}

// TerminateAuction implements the MsgServer interface for terminating Dutch auctions
func (s msgServer) TerminateAuction(goCtx context.Context, msg *types.MsgTerminateAuction) (*types.MsgTerminateAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != s.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", s.authority, msg.Authority)
	}

	// FIXME: Terminate the auction
	_ = ctx

	return &types.MsgTerminateAuctionResponse{}, nil
}
