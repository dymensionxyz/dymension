package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
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

// UpdateParams defines a method to update the module params
func (ms msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if ms.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := ms.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// SetAcceptedTokens allows authority to set/update accepted tokens for auctions
func (ms msgServer) SetAcceptedTokens(goCtx context.Context, req *types.MsgSetAcceptedTokens) (*types.MsgSetAcceptedTokensResponse, error) {
	if ms.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	for _, token := range req.AcceptedTokens {
		denoms, err := ms.Keeper.ammKeeper.GetPoolDenoms(ctx, token.PoolId)
		if err != nil {
			return nil, err
		}

		if len(denoms) != 2 {
			return nil, fmt.Errorf("pool must have two denoms")
		}
		if (denoms[0] != ms.Keeper.baseDenom && denoms[1] != token.Token) ||
			(denoms[1] != ms.Keeper.baseDenom && denoms[0] != token.Token) {
			return nil, fmt.Errorf("pool must have the token denom and the base denom, got %s and %s", denoms[0], denoms[1])
		}
	}

	if err := ms.Keeper.SetAcceptedTokens(ctx, req.AcceptedTokens); err != nil {
		return nil, err
	}

	return &types.MsgSetAcceptedTokensResponse{}, nil
}

// Buy handles token purchase requests
func (ms msgServer) Buy(goCtx context.Context, req *types.MsgBuy) (*types.MsgBuyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert buyer address
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, err.Error())
	}

	paymentCoin, err := ms.Keeper.Buy(ctx, buyer, req.AuctionId, req.AmountToBuy, req.DenomToPay)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{
		TokensPurchased: req.AmountToBuy,
		PaymentCoin:     paymentCoin,
	}, nil
}

func (ms msgServer) BuyExactSpend(goCtx context.Context, req *types.MsgBuyExactSpend) (*types.MsgBuyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert buyer address
	buyer, err := sdk.AccAddressFromBech32(req.Buyer)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, err.Error())
	}

	tokensPurchased, err := ms.Keeper.BuyExactSpend(ctx, buyer, req.AuctionId, req.PaymentCoin)
	if err != nil {
		return nil, err
	}

	return &types.MsgBuyResponse{
		TokensPurchased: tokensPurchased,
		PaymentCoin:     req.PaymentCoin,
	}, nil
}

// ClaimTokens handles token claim requests
func (ms msgServer) ClaimTokens(goCtx context.Context, req *types.MsgClaimTokens) (*types.MsgClaimTokensResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Convert claimer address
	claimer, err := sdk.AccAddressFromBech32(req.Claimer)
	if err != nil {
		return nil, errorsmod.Wrap(types.ErrInvalidAddress, err.Error())
	}

	claimedAmount, err := ms.ClaimVestedTokens(ctx, claimer, req.AuctionId)
	if err != nil {
		return nil, err
	}

	return &types.MsgClaimTokensResponse{
		ClaimedAmount: claimedAmount,
	}, nil
}
