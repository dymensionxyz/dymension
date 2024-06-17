package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
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

func (m msgServer) FulfillOrder(goCtx context.Context, msg *types.MsgFulfillOrder) (*types.MsgFulfillOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger()

	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	demandOrder, err := m.GetOutstandingOrder(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	// Check that the fulfiller expected fee is equal to the demand order fee
	for _, coin := range demandOrder.Fee {
		expectedFee, _ := sdk.NewIntFromString(msg.ExpectedFee)
		if !coin.Amount.Equal(expectedFee) {
			return nil, types.ErrExpectedFeeNotMet
		}
	}

	// Check that the fulfiller has enough balance to fulfill the order
	fulfillerAccount := m.ak.GetAccount(ctx, msg.GetFulfillerBech32Address())
	if fulfillerAccount == nil {
		return nil, types.ErrFulfillerAddressDoesNotExist
	}

	// Send the funds from the fulfiller to the eibc packet original recipient
	err = m.bk.SendCoins(ctx, fulfillerAccount.GetAddress(), demandOrder.GetRecipientBech32Address(), demandOrder.Price)
	if err != nil {
		logger.Error("Failed to send coins", "error", err)
		return nil, err
	}
	// Fulfill the order by updating the order status and underlying packet recipient
	err = m.Keeper.SetOrderFulfilled(ctx, demandOrder, fulfillerAccount.GetAddress())
	if err != nil {
		return nil, err
	}

	return &types.MsgFulfillOrderResponse{}, err
}

// UpdateDemandOrder implements types.MsgServer.
func (m msgServer) UpdateDemandOrder(goCtx context.Context, msg *types.MsgUpdateDemandOrder) (*types.MsgUpdateDemandOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// Check that the order exists in status PENDING
	demandOrder, err := m.GetOutstandingOrder(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	// Check that the signer is the order owner
	orderOwner := demandOrder.GetRecipientBech32Address()
	msgSigner := msg.GetSignerAddr()
	if !msgSigner.Equals(orderOwner) {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only the recipient can update the order")
	}

	raPacket, err := m.dack.GetRollappPacket(ctx, demandOrder.TrackingPacketKey)
	if err != nil {
		return nil, err
	}

	var data transfertypes.FungibleTokenPacketData
	if err := transfertypes.ModuleCdc.UnmarshalJSON(raPacket.GetPacket().GetData(), &data); err != nil {
		return nil, err
	}

	// Get the bridging fee multiplier
	// ErrAck or Timeout packets do not incur bridging fees
	bridgingFeeMultiplier := m.dack.BridgingFee(ctx)
	raPacketType := raPacket.GetType()
	if raPacketType != commontypes.RollappPacket_ON_RECV {
		bridgingFeeMultiplier = sdk.ZeroDec()
	}

	// calculate the new price: transferTotal - newFee - bridgingFee
	newFeeInt, _ := sdk.NewIntFromString(msg.NewFee)
	transferTotal, _ := sdk.NewIntFromString(data.Amount)
	newPrice, err := types.CalcPriceWithBridgingFee(transferTotal, newFeeInt, bridgingFeeMultiplier)
	if err != nil {
		return nil, err
	}

	denom := demandOrder.Price[0].Denom
	demandOrder.Fee = sdk.NewCoins(sdk.NewCoin(denom, newFeeInt))
	demandOrder.Price = sdk.NewCoins(sdk.NewCoin(denom, newPrice))

	err = m.SetDemandOrder(ctx, demandOrder)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateDemandOrderResponse{}, nil
}

func (m msgServer) GetOutstandingOrder(ctx sdk.Context, orderId string) (*types.DemandOrder, error) {
	// Check that the order exists in status PENDING
	demandOrder, err := m.GetDemandOrder(ctx, commontypes.Status_PENDING, orderId)
	if err != nil {
		return nil, err
	}

	return demandOrder, demandOrder.ValidateOrderIsOutstanding()
}
