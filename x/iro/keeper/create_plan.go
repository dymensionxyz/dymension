package keeper

import (
	"context"
	"errors"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// CreatePlan implements types.MsgServer.
func (m msgServer) CreatePlan(goCtx context.Context, req *types.MsgCreatePlan) (*types.MsgCreatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rollapp, found := m.Keeper.rk.GetRollapp(ctx, req.RollappId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "rollapp not found")
	}

	if rollapp.Owner != req.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	// Validate rollapp preconditions
	err := ValidateRollappPreconditions(rollapp)
	if err != nil {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, err)
	}

	// validate end time is in the future
	if req.EndTime.Before(ctx.BlockTime()) {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidEndTime)
	}

	// Check if the plan already exists
	_, found = m.Keeper.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrPlanExists)
	}

	planId, err := m.Keeper.CreatePlan(ctx, req.AllocatedAmount, req.StartTime, req.EndTime, rollapp, req.BondingCurve)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePlanResponse{
		PlanId: planId,
	}, nil

}

func ValidateRollappPreconditions(rollapp rollapptypes.Rollapp) error {
	if rollapp.GenesisChecksum == "" {
		return types.ErrRollappGenesisChecksumNotSet
	}

	if rollapp.Metadata.TokenSymbol == "" {
		return types.ErrRollappTokenSymbolNotSet
	}

	// rollapp cannot be sealed when creating a plan
	if rollapp.Sealed {
		return types.ErrRollappSealed
	}

	return nil
}

func (k Keeper) CreatePlan(ctx sdk.Context, allocatedAmount math.Int, start, end time.Time, rollapp rollapptypes.Rollapp, curve types.BondingCurve) (string, error) {

	// FIXME: create a module account for the plan

	// FIXME: get decimals from the caller
	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.Metadata.TokenSymbol, 18)
	if err != nil {
		return "", err
	}

	// FIXME: move the minted tokens to plan’s module account

	// FIXME: set curve config + validation

	// Create a new plan
	plan := types.Plan{
		Id:               k.GetLastPlanId(ctx) + 1,
		RollappId:        rollapp.RollappId,
		ModuleAccAddress: "", // FIXME
		TotalAllocation:  allocation,
		BondingCurve:     curve,
		StartTime:        start,
		EndTime:          end,
		SoldAmt:          math.Int{},
		ClaimedAmt:       math.Int{},
	}

	// Set the plan in the store
	k.SetPlan(ctx, plan)
	k.SetLastPlanId(ctx, plan.Id)

	return fmt.Sprintf("%d", plan.Id), nil
}

// MintAllocation mints the allocated amount and registers the denom in the bank denom metadata store
func (k Keeper) MintAllocation(ctx sdk.Context, allocatedAmount math.Int, rollappId, rollappSymbolName string, exponent uint64) (sdk.Coin, error) {
	// Register the denom in the bank denom metadata store
	baseDenom := fmt.Sprintf("FUT_%s", rollappId)
	displayDenom := fmt.Sprintf("FUT_%s", rollappSymbolName)
	metadata := banktypes.Metadata{
		Description: fmt.Sprintf("Future token for rollapp %s", rollappId),
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: baseDenom, Exponent: 0, Aliases: []string{}},
			{Denom: displayDenom, Exponent: uint32(exponent), Aliases: []string{}},
		},
		Base:    baseDenom,
		Name:    baseDenom,
		Display: displayDenom,
		Symbol:  displayDenom,
	}
	if err := metadata.Validate(); err != nil {
		return sdk.Coin{}, errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), fmt.Sprintf("metadata: %v", metadata))
	}

	if k.bk.HasDenomMetaData(ctx, baseDenom) {
		return sdk.Coin{}, errors.New("denom already exists")
	}
	k.bk.SetDenomMetaData(ctx, metadata)

	toMint := sdk.NewCoin(baseDenom, allocatedAmount)
	err := k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(toMint))
	if err != nil {
		return sdk.Coin{}, err
	}
	return toMint, nil
}