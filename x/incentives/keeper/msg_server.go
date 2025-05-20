package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"

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

// UpdateParams implements types.MsgServer.
func (m msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Check if the sender is the authority
	if req.Authority != m.keeper.authority {
		return nil, errorsmod.Wrap(sdkerrors.ErrUnauthorized, "only the gov module can update params")
	}

	err := req.Params.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// TODO: make sure MinValueForDistribution is same as txfees basedenom

	m.keeper.SetParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}

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
	if err = server.keeper.ChargeGaugesFee(ctx, owner, fee); err != nil {
		return nil, fmt.Errorf("charge gauge fee: %w", err)
	}

	var gaugeID uint64
	switch distr := msg.DistributeTo.(type) {
	case *types.MsgCreateGauge_Asset:
		gaugeID, err = server.keeper.CreateAssetGauge(ctx, msg.IsPerpetual, owner, msg.Coins, *distr.Asset, msg.StartTime, msg.NumEpochsPaidOver)
		if err != nil {
			return nil, fmt.Errorf("create gauge: %w", err)
		}
	case *types.MsgCreateGauge_Endorsement:
		gaugeID, err = server.keeper.CreateEndorsementGauge(ctx, msg.IsPerpetual, owner, msg.Coins, *distr.Endorsement, msg.StartTime, msg.NumEpochsPaidOver)
		if err != nil {
			return nil, fmt.Errorf("create gauge: %w", err)
		}
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
	if err = server.keeper.ChargeGaugesFee(ctx, owner, fee); err != nil {
		return nil, fmt.Errorf("charge gauge fee: %w", err)
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

// chargeGaugesFee deducts a fee in the base denom from the specified address.
// The fee is charged from the payer and sent to x/txfees to be burned.
func (k Keeper) ChargeGaugesFee(ctx sdk.Context, payer sdk.AccAddress, fee math.Int) (err error) {
	feeDenom, err := k.tk.GetBaseDenom(ctx)
	if err != nil {
		return err
	}

	return k.tk.ChargeFeesFromPayer(ctx, payer, sdk.NewCoin(feeDenom, fee), nil)
}
