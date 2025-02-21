package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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

	err := msg.ValidateBasic() // TODO: remove, sdk does this
	if err != nil {
		return nil, err
	}

	demandOrder, err := m.GetOutstandingOrder(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	// Check that the fulfiller expected fee is equal to the demand order fee
	expectedFee, _ := math.NewIntFromString(msg.ExpectedFee)
	orderFee := demandOrder.GetFeeAmount()
	if !orderFee.Equal(expectedFee) {
		return nil, types.ErrExpectedFeeNotMet
	}

	err = m.Fulfill(ctx, demandOrder, msg.GetFulfillerBech32Address())
	if err != nil {
		logger.Error("Fulfill order", "error", err)
		return nil, err
	}

	return &types.MsgFulfillOrderResponse{}, nil
}

func (m msgServer) FulfillOrderAuthorized(goCtx context.Context, msg *types.MsgFulfillOrderAuthorized) (*types.MsgFulfillOrderAuthorizedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	logger := ctx.Logger()

	err := msg.ValidateBasic() // TODO: remove, sdk does this
	if err != nil {
		return nil, err
	}

	demandOrder, err := m.GetOutstandingOrder(ctx, msg.OrderId)
	if err != nil {
		return nil, err
	}

	// check compat between the fulfillment and current order and packet status
	if err := m.validateOrder(demandOrder, msg, ctx); err != nil {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, err.Error())
	}

	lpAccount := m.ak.GetAccount(ctx, msg.GetLPBech32Address())
	if lpAccount == nil {
		return nil, types.ErrLPAccountDoesNotExist
	}

	// Send the funds from the lpAccount to the eibc packet original recipient
	err = m.bk.SendCoins(ctx, lpAccount.GetAddress(), demandOrder.GetRecipientBech32Address(), demandOrder.Price)
	if err != nil {
		logger.Error("Failed to send price to recipient", "error", err)
		return nil, err
	}

	operatorAccount := m.ak.GetAccount(ctx, msg.GetOperatorFeeBech32Address())
	if operatorAccount == nil {
		return nil, types.ErrOperatorFeeAccountDoesNotExist
	}

	fee := math.LegacyNewDecFromInt(demandOrder.GetFeeAmount())
	operatorFee := fee.MulTruncate(msg.OperatorFeeShare).TruncateInt()

	if operatorFee.IsPositive() {
		// Send the fee part to the fulfiller/operator
		err = m.bk.SendCoins(ctx, lpAccount.GetAddress(), operatorAccount.GetAddress(), sdk.NewCoins(sdk.NewCoin(demandOrder.Price[0].Denom, operatorFee)))
		if err != nil {
			logger.Error("Failed to send fee part to operator", "error", err)
			return nil, err
		}
	}

	if err = m.Keeper.SetOrderFulfilled(ctx, demandOrder, operatorAccount.GetAddress(), lpAccount.GetAddress()); err != nil {
		return nil, err
	}

	if err = uevent.EmitTypedEvent(ctx, demandOrder.GetFulfilledAuthorizedEvent(
		demandOrder.CreationHeight,
		msg.LpAddress,
		operatorAccount.GetAddress().String(),
		operatorFee.String(),
	)); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgFulfillOrderAuthorizedResponse{}, nil
}

// TODO: rename and fix signature (ctx first)
func (m msgServer) validateOrder(demandOrder *types.DemandOrder, msg *types.MsgFulfillOrderAuthorized, ctx sdk.Context) error {
	if demandOrder.RollappId != msg.RollappId {
		return types.ErrRollappIdMismatch
	}

	if !demandOrder.Price.Equal(msg.Price) {
		return types.ErrPriceMismatch
	}

	// Check that the expected fee is equal to the demand order fee
	expectedFee, _ := math.NewIntFromString(msg.ExpectedFee)
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

	// TODO: extract to rollapp keeper func HaveHeight(..)

	// as it is not currently possible to make IBC transfers without a canonical client,
	// we can assume that there has to exist at least one state info record for the rollapp
	stateInfo, ok := m.rk.GetLatestStateInfo(ctx, demandOrder.RollappId)
	if !ok {
		return false, types.ErrRollappStateInfoNotFound
	}

	lastHeight := stateInfo.GetLatestHeight()

	return raPacket.ProofHeight <= lastHeight, nil
}

// UpdateDemandOrder implements types.MsgServer.
func (m msgServer) UpdateDemandOrder(goCtx context.Context, msg *types.MsgUpdateDemandOrder) (*types.MsgUpdateDemandOrderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.ValidateBasic() // TODO: remove, sdk does this
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
		// TODO: isn't this internal error?
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
		bridgingFeeMultiplier = math.LegacyZeroDec()
	}

	// calculate the new price: transferTotal - newFee - bridgingFee
	newFeeInt, _ := math.NewIntFromString(msg.NewFee)
	transferTotal, _ := math.NewIntFromString(data.Amount)
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

	if err = uevent.EmitTypedEvent(ctx, demandOrder.GetUpdatedEvent(raPacket.ProofHeight, data.Amount)); err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return &types.MsgUpdateDemandOrderResponse{}, nil
}

func (m msgServer) FindFulfiller(goCtx context.Context, msg *types.MsgFindFulfiller) (*types.MsgFindFulfillerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.ValidateBasic()
	if err != nil {
		return nil, errorsmod.Wrap(err, "vbasic")
	}

	return &types.MsgFindFulfillerResponse{}, m.Keeper.FulfillByOnDemandLP(ctx, msg.OrderId, msg.Rng)
}

func (m msgServer) CreateOnDemandLP(goCtx context.Context, msg *types.MsgCreateOnDemandLP) (*types.MsgCreateOnDemandLPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.ValidateBasic()
	if err != nil {
		return nil, errorsmod.Wrap(err, "vbasic")
	}

	id, err := m.Keeper.CreateLP(ctx, msg.Lp)
	if err != nil {
		return nil, errorsmod.Wrap(err, "create lp")
	}

	return &types.MsgCreateOnDemandLPResponse{Id: id}, nil

}

func (m msgServer) DeleteOnDemandLP(goCtx context.Context, msg *types.MsgDeleteOnDemandLP) (*types.MsgDeleteOnDemandLPResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.ValidateBasic()
	if err != nil {
		return nil, errorsmod.Wrap(err, "vbasic")
	}

	for _, id := range msg.Ids {
		err := m.Keeper.DeleteLP(ctx, msg.MustAcc(), id, "user request")
		if err != nil {
			return nil, errorsmod.Wrapf(err, "delete id: %d", id)
		}
	}

	return &types.MsgDeleteOnDemandLPResponse{}, nil
}
