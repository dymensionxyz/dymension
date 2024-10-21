package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

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
	expectedFee, _ := sdk.NewIntFromString(msg.ExpectedFee)
	orderFee := demandOrder.GetFeeAmount()
	if !orderFee.Equal(expectedFee) {
		return nil, types.ErrExpectedFeeNotMet
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
	if err = m.Keeper.SetOrderFulfilled(ctx, demandOrder, fulfillerAccount.GetAddress(), nil); err != nil {
		return nil, err
	}

	if err = uevent.EmitTypedEvent(ctx, demandOrder.GetFulfilledEvent()); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFulfillOrderResponse{}, nil
}

func (m msgServer) FulfillOrderAuthorized(goCtx context.Context, msg *types.MsgFulfillOrderAuthorized) (*types.MsgFulfillOrderAuthorizedResponse, error) {
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

	if err := m.validateOrder(demandOrder, msg, ctx); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	}

	lpAccount := m.ak.GetAccount(ctx, msg.GetLPBech32Address())
	if lpAccount == nil {
		return nil, types.ErrGranterAddressDoesNotExist
	}

	// Send the funds from the lpAccount to the eibc packet original recipient
	err = m.bk.SendCoins(ctx, lpAccount.GetAddress(), demandOrder.GetRecipientBech32Address(), demandOrder.Price)
	if err != nil {
		logger.Error("Failed to send price to recipient", "error", err)
		return nil, err
	}

	// TODO: will this work for Policy address?
	operatorAccount := m.ak.GetAccount(ctx, msg.GetOperatorBech32Address())
	if operatorAccount == nil {
		return nil, types.ErrFulfillerAddressDoesNotExist
	}

	// by default, the operator account receives the operator share
	feePartReceiver := operatorAccount
	// if the operator fee address is provided, the operator fee share is sent to that address
	if msg.OperatorFeeAddress != "" {
		operatorFeeAccount := m.ak.GetAccount(ctx, msg.GetOperatorFeeBech32Address())
		if operatorFeeAccount == nil {
			return nil, types.ErrOperatorAddressDoesNotExist
		}
		feePartReceiver = operatorFeeAccount
	}

	fee := sdk.NewDecFromInt(demandOrder.GetFeeAmount())
	operatorFee := fee.Mul(msg.OperatorFeeShare.Dec).TruncateInt()

	if operatorFee.IsPositive() {
		// Send the fee part to the fulfiller/operator
		err = m.bk.SendCoins(ctx, lpAccount.GetAddress(), feePartReceiver.GetAddress(), sdk.NewCoins(sdk.NewCoin(demandOrder.Price[0].Denom, operatorFee)))
		if err != nil {
			logger.Error("Failed to send fee part to operator", "error", err)
			return nil, err
		}
	}

	if err = m.Keeper.SetOrderFulfilled(ctx, demandOrder, operatorAccount.GetAddress(), lpAccount.GetAddress()); err != nil {
		return nil, err
	}

	if err = uevent.EmitTypedEvent(ctx, demandOrder.GetFulfilledEvent()); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFulfillOrderAuthorizedResponse{}, nil
}

func (m msgServer) validateOrder(demandOrder *types.DemandOrder, msg *types.MsgFulfillOrderAuthorized, ctx sdk.Context) error {
	if demandOrder.RollappId != msg.RollappId {
		return types.ErrRollappIdMismatch
	}

	if !demandOrder.Price.IsEqual(msg.Price) {
		return types.ErrPriceMismatch
	}

	// Check that the expected fee is equal to the demand order fee
	expectedFee, _ := sdk.NewIntFromString(msg.ExpectedFee)
	orderFee := demandOrder.GetFeeAmount()
	if !orderFee.Equal(expectedFee) {
		return types.ErrExpectedFeeNotMet
	}

	if msg.SettlementValidated {
		validated, err := m.checkIfSettlementValidated(ctx, demandOrder)
		if err != nil {
			return fmt.Errorf("check if settlement validated: %w", err)
		}

		if !validated {
			return types.ErrOrderNotSettlementValidated
		}
	}
	return nil
}

func (m msgServer) checkIfSettlementValidated(ctx sdk.Context, demandOrder *types.DemandOrder) (bool, error) {
	raPacket, err := m.dack.GetRollappPacket(ctx, demandOrder.TrackingPacketKey)
	if err != nil {
		return false, fmt.Errorf("get rollapp packet: %w", err)
	}

	// as it is not currently possible to make IBC transfers without a canonical client,
	// we can assume that there has to exist at least one state info record for the rollapp
	stateInfo, ok := m.rk.GetLatestStateInfo(ctx, demandOrder.RollappId)
	if !ok {
		return false, types.ErrRollappStateInfoNotFound
	}

	lastHeight := stateInfo.GetLatestHeight()

	if lastHeight < raPacket.ProofHeight {
		return false, nil
	}

	return true, nil
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

	if err = m.SetDemandOrder(ctx, demandOrder); err != nil {
		return nil, err
	}

	if err = uevent.EmitTypedEvent(ctx, demandOrder.GetUpdatedEvent()); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
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
