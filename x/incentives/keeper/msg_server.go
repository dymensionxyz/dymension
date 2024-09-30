package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
)

// msgServer provides a way to reference keeper pointer in the message server interface.
type msgServer struct {
	keeper *Keeper
}

// NewMsgServerImpl returns an instance of MsgServer for the provided keeper.
func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{
		keeper: keeper,
	}
}

var _ types.MsgServer = msgServer{}

// CreateGauge creates a gauge and sends coins to the gauge.
// Creation fee is charged from the address and sent to the txfees module to be burned.
// Emits create gauge event and returns the create gauge response.
func (server msgServer) CreateGauge(goCtx context.Context, msg *types.MsgCreateGauge) (*types.MsgCreateGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}
	// Charge fess based on the number of coins to add
	// Fee = CreateGaugeBaseFee + AddDenomFee * NumDenoms
	params := server.keeper.GetParams(ctx)
	fee := params.CreateGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(msg.Coins))))
	if err = server.keeper.ChargeGaugesFee(ctx, owner, fee, msg.Coins); err != nil {
		return nil, err
	}

	gaugeID, err := server.keeper.CreateGauge(ctx, msg.IsPerpetual, owner, msg.Coins, msg.DistributeTo, msg.StartTime, msg.NumEpochsPaidOver)
	if err != nil {
		return nil, fmt.Errorf("create gauge: %w", err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtCreateGauge,
			sdk.NewAttribute(types.AttributeGaugeID, osmoutils.Uint64ToString(gaugeID)),
		),
	})

	return &types.MsgCreateGaugeResponse{}, nil
}

// AddToGauge adds coins to gauge.
// Emits add to gauge event and returns the add to gauge response.
func (server msgServer) AddToGauge(goCtx context.Context, msg *types.MsgAddToGauge) (*types.MsgAddToGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	gauge, err := server.keeper.GetGaugeByID(ctx, msg.GaugeId)
	if err != nil {
		return nil, err
	}

	// Charge fess based on the number of coins to add
	// Fee = AddToGaugeBaseFee + AddDenomFee * (NumAddedDenoms + NumGaugeDenoms)
	params := server.keeper.GetParams(ctx)
	fee := params.AddToGaugeBaseFee.Add(params.AddDenomFee.MulRaw(int64(len(msg.Rewards) + len(gauge.Coins))))
	if err = server.keeper.ChargeGaugesFee(ctx, owner, fee, msg.Rewards); err != nil {
		return nil, err
	}

	err = server.keeper.AddToGaugeRewards(ctx, owner, msg.Rewards, gauge)
	if err != nil {
		return nil, fmt.Errorf("add to gauge rewards: %w", err)
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToGauge,
			sdk.NewAttribute(types.AttributeGaugeID, osmoutils.Uint64ToString(msg.GaugeId)),
		),
	})

	return &types.MsgAddToGaugeResponse{}, nil
}

// ChargeGaugesFee charges fee in the base denom on the address if the address has
// balance that is less than fee + amount of the coin from gaugeCoins that is of base denom.
// gaugeCoins might not have a coin of tx base denom. In that case, fee is only compared to balance.
// The fee is sent to the txfees module, to be burned.
func (k Keeper) ChargeGaugesFee(ctx sdk.Context, address sdk.AccAddress, fee sdk.Int, gaugeCoins sdk.Coins) (err error) {
	var feeDenom string
	if k.tk == nil {
		feeDenom, err = sdk.GetBaseDenom()
	} else {
		feeDenom, err = k.tk.GetBaseDenom(ctx)
	}
	if err != nil {
		return err
	}

	totalCost := gaugeCoins.AmountOf(feeDenom).Add(fee)
	accountBalance := k.bk.GetBalance(ctx, address, feeDenom).Amount

	if accountBalance.LT(totalCost) {
		return errorsmod.Wrapf(sdkerrors.ErrInsufficientFunds, "account's balance is less than the total cost of the message. Balance: %s %s, Total Cost: %s", feeDenom, accountBalance, totalCost)
	}

	return k.bk.SendCoinsFromAccountToModule(ctx, address, txfeestypes.ModuleName, sdk.NewCoins(sdk.NewCoin(feeDenom, fee)))
}
