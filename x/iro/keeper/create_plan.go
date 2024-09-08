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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// This function is used to create a new plan for a rollapp.
// Validations on the request:
// - The rollapp must exist, with no IRO plan
// - The rollapp must be owned by the creator of the plan
// - The rollapp PreLaunchTime must be in the future
// - The plan duration must be at least the minimum duration set in the module params
// - The incentive plan params must be valid and meet the minimum requirements set in the module params
func (m msgServer) CreatePlan(goCtx context.Context, req *types.MsgCreatePlan) (*types.MsgCreatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rollapp, found := m.Keeper.rk.GetRollapp(ctx, req.RollappId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "rollapp not found")
	}

	if rollapp.Owner != req.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	// validate end time is in the future
	if req.PreLaunchTime.Before(ctx.BlockTime()) {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidEndTime)
	}

	startTime := req.StartTime
	if startTime.Before(ctx.BlockTime()) {
		startTime = ctx.BlockTime()
	}
	// check minimal duration
	if startTime.Add(m.Keeper.GetParams(ctx).MinPlanDuration).After(req.PreLaunchTime) {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidEndTime)
	}

	// validate incentive plan params
	incentivesMinParams := m.Keeper.GetParams(ctx).IncentivePlanMinimumParams
	if req.IncentivePlanParams.NumEpochsPaidOver < incentivesMinParams.NumEpochsPaidOver {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "num epochs paid over"))
	}
	if req.IncentivePlanParams.StartTimeAfterSettlement < incentivesMinParams.StartTimeAfterSettlement {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "start time after settlement"))
	}

	// Check if the plan already exists
	_, found = m.Keeper.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrPlanExists)
	}

	planId, err := m.Keeper.CreatePlan(ctx, req.AllocatedAmount, startTime, req.PreLaunchTime, rollapp, req.BondingCurve, req.IncentivePlanParams)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePlanResponse{
		PlanId: planId,
	}, nil
}

// ValidateRollappPreconditions validates the preconditions for creating a plan
// - GenesisInfo fields must be set
// - Rollapp must not be Launched
func ValidateRollappPreconditions(rollapp rollapptypes.Rollapp) error {
	if !rollapp.GenesisInfoFieldsAreSet() {
		return types.ErrRollappGenesisInfoNotSet
	}

	// rollapp cannot be launched when creating a plan
	if rollapp.Launched {
		return types.ErrRollappSealed
	}

	return nil
}

// CreatePlan creates a new IRO plan for a rollapp
func (k Keeper) CreatePlan(ctx sdk.Context, allocatedAmount math.Int, start, preLaunchTime time.Time, rollapp rollapptypes.Rollapp, curve types.BondingCurve, incentivesParams types.IncentivePlanParams) (string, error) {
	err := ValidateRollappPreconditions(rollapp)
	if err != nil {
		return "", errors.Join(gerrc.ErrFailedPrecondition, err)
	}

	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Display, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	if err != nil {
		return "", err
	}

	plan := types.NewPlan(k.GetNextPlanIdAndIncrement(ctx), rollapp.RollappId, allocation, curve, start, preLaunchTime, incentivesParams)
	// Create a new module account for the IRO plan
	_, err = k.CreateModuleAccountForPlan(ctx, plan)
	if err != nil {
		return "", err
	}

	// charge creation fee
	fee := sdk.NewCoin(appparams.BaseDenom, k.GetParams(ctx).CreationFee)
	err = k.BK.SendCoins(ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), plan.GetAddress(), sdk.NewCoins(fee))
	if err != nil {
		return "", err
	}

	// Set the plan in the store
	k.SetPlan(ctx, plan)

	// Update the rollapp with the IRO plan pre launch time. This will also seals the genesis info
	k.rk.UpdateRollappWithIROPlanAndSeal(ctx, rollapp.RollappId, preLaunchTime)

	return fmt.Sprintf("%d", plan.Id), nil
}

func (k Keeper) CreateModuleAccountForPlan(ctx sdk.Context, plan types.Plan) (authtypes.ModuleAccountI, error) {
	moduleAccount := authtypes.NewEmptyModuleAccount(plan.ModuleAccName())
	moduleAccountI, ok := (k.AK.NewAccount(ctx, moduleAccount)).(authtypes.ModuleAccountI)
	if !ok {
		return nil, errorsmod.Wrap(gerrc.ErrInternal, "failed to create module account")
	}
	k.AK.SetModuleAccount(ctx, moduleAccountI)
	return moduleAccountI, nil
}

// MintAllocation mints the allocated amount and registers the denom in the bank denom metadata store
func (k Keeper) MintAllocation(ctx sdk.Context, allocatedAmount math.Int, rollappId, rollappTokenSymbol string, exponent uint64) (sdk.Coin, error) {
	baseDenom := fmt.Sprintf("%s_%s", types.IROTokenPrefix, rollappId)
	displayDenom := fmt.Sprintf("%s_%s", types.IROTokenPrefix, rollappTokenSymbol)
	metadata := banktypes.Metadata{
		Description: fmt.Sprintf("Future token for rollapp %s", rollappId),
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: baseDenom, Exponent: 0, Aliases: []string{}},
			{Denom: displayDenom, Exponent: uint32(exponent), Aliases: []string{}}, //nolint:gosec
		},
		Base:    baseDenom,
		Name:    baseDenom,
		Display: displayDenom,
		Symbol:  displayDenom,
	}
	if err := metadata.Validate(); err != nil {
		return sdk.Coin{}, errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), fmt.Sprintf("metadata: %v", metadata))
	}

	if k.BK.HasDenomMetaData(ctx, baseDenom) {
		return sdk.Coin{}, errors.New("denom already exists")
	}
	k.BK.SetDenomMetaData(ctx, metadata)

	minted := sdk.NewCoin(baseDenom, allocatedAmount)
	err := k.BK.MintCoins(ctx, types.ModuleName, sdk.NewCoins(minted))
	if err != nil {
		return sdk.Coin{}, err
	}
	return minted, nil
}
