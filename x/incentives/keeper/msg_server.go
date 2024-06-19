package keeper

import (
	"context"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
// Emits create gauge event and returns the create gauge response.
func (server msgServer) CreateGauge(goCtx context.Context, msg *types.MsgCreateGauge) (*types.MsgCreateGaugeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	if err := server.keeper.chargeFeeIfSufficientFeeDenomBalance(ctx, owner, types.CreateGaugeFee, msg.Coins); err != nil {
		return nil, err
	}

	gaugeID, err := server.keeper.CreateGauge(ctx, msg.IsPerpetual, owner, msg.Coins, msg.DistributeTo, msg.StartTime, msg.NumEpochsPaidOver)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
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

	if err := server.keeper.chargeFeeIfSufficientFeeDenomBalance(ctx, owner, types.AddToGaugeFee, msg.Rewards); err != nil {
		return nil, err
	}
	err = server.keeper.AddToGaugeRewards(ctx, owner, msg.Rewards, msg.GaugeId)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, err.Error())
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.TypeEvtAddToGauge,
			sdk.NewAttribute(types.AttributeGaugeID, osmoutils.Uint64ToString(msg.GaugeId)),
		),
	})

	return &types.MsgAddToGaugeResponse{}, nil
}

// chargeFeeIfSufficientFeeDenomBalance charges fee in the base denom on the address if the address has
// balance that is less than fee + amount of the coin from gaugeCoins that is of base denom.
// gaugeCoins might not have a coin of tx base denom. In that case, fee is only compared to balance.
// The fee is sent to the community pool.
// Returns nil on success, error otherwise.
func (k Keeper) chargeFeeIfSufficientFeeDenomBalance(ctx sdk.Context, address sdk.AccAddress, fee sdk.Int, gaugeCoins sdk.Coins) (err error) {
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
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "account's balance of %s (%s) is less than the total cost of the message (%s)", feeDenom, accountBalance, totalCost)
	}

	if err := k.ck.FundCommunityPool(ctx, sdk.NewCoins(sdk.NewCoin(feeDenom, fee)), address); err != nil {
		return err
	}
	return nil
}
