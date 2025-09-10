package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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

	k := s.Keeper

	moduleBalance := k.bk.GetAllBalances(ctx, authtypes.NewModuleAddress(types.ModuleName))
	alreadyAllocatedCoins := k.GetModuleToDistributeCoins(ctx)

	if !sdk.NewCoins(msg.Allocation).IsAllLTE(moduleBalance.Sub(alreadyAllocatedCoins...)) {
		return nil, fmt.Errorf("insufficient module balance to distribute coins")
	}

	// move funds from the streamer account to the auction account
	err := k.bk.SendCoinsFromModuleToModule(ctx, types.ModuleName, hatypes.ModuleName, sdk.NewCoins(msg.Allocation))
	if err != nil {
		return nil, err
	}

	// Create the auction and stream
	auctionID, err := s.Keeper.otck.CreateAuction(
		ctx,
		msg.Allocation,
		msg.StartTime,
		msg.EndTime,
		msg.InitialDiscount,
		msg.MaxDiscount,
		hatypes.Auction_VestingParams{
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
