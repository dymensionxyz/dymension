package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	k := server.keeper
	owner, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return nil, err
	}

	// charge creation fee
	if err := k.chargeCreationFee(ctx, owner); err != nil {
		return nil, fmt.Errorf("charge creation fee: %w", err)
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

	err = server.keeper.AddToGaugeRewards(ctx, owner, msg.Rewards, msg.GaugeId)
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

// chargeCreationFee charges the creation fee from the address.
// The fee is sent to the txfees module, to be burned.
func (k Keeper) chargeCreationFee(ctx sdk.Context, address sdk.AccAddress) (err error) {
	feeDenom, err := k.tk.GetBaseDenom(ctx)
	if err != nil {
		return errorsmod.Wrapf(gerrc.ErrInternal, "get base denom: %v", err)
	}
	return k.bk.SendCoinsFromAccountToModule(ctx, address, txfeestypes.ModuleName, sdk.NewCoins(sdk.NewCoin(feeDenom, types.CreateGaugeFee)))
}
