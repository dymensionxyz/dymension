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
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// This function is used to create a new plan for a rollapp.
// Non stateful validation happens on the req.ValidateBasic() method
// Stateful validations on the request:
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

	startTime := req.StartTime
	if startTime.Before(ctx.BlockTime()) {
		startTime = ctx.BlockTime()
	}
	// check minimal plan duration
	if req.IroPlanDuration < m.Keeper.GetParams(ctx).MinPlanDuration {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidEndTime)
	}
	preLaunchTime := startTime.Add(req.IroPlanDuration)

	// check minimal liquidity part
	if req.LiquidityPart.LT(m.Keeper.GetParams(ctx).MinLiquidityPart) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, fmt.Sprintf("liquidity part must be at least %s", m.Keeper.GetParams(ctx).MinLiquidityPart))
	}

	// check vesting params
	if req.VestingDuration < m.Keeper.GetParams(ctx).MinVestingDuration {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, fmt.Sprintf("vesting duration must be at least %s", m.Keeper.GetParams(ctx).MinVestingDuration))
	}

	if req.VestingStartTimeAfterSettlement < m.Keeper.GetParams(ctx).MinVestingStartTimeAfterSettlement {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, fmt.Sprintf("vesting start time after settlement must be at least %s", m.Keeper.GetParams(ctx).MinVestingStartTimeAfterSettlement))
	}

	// validate incentive plan params
	params := m.Keeper.GetParams(ctx)
	if req.IncentivePlanParams.NumEpochsPaidOver < params.IncentivesMinNumEpochsPaidOver {
		return nil, errors.Join(gerrc.ErrInvalidArgument, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "num epochs paid over"))
	}
	if req.IncentivePlanParams.StartTimeAfterSettlement < params.IncentivesMinStartTimeAfterSettlement {
		return nil, errors.Join(gerrc.ErrInvalidArgument, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "start time after settlement"))
	}

	// Check if the plan already exists
	_, found = m.Keeper.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrPlanExists)
	}

	found = false
	for _, gAcc := range rollapp.GenesisInfo.Accounts() {
		if gAcc.Address == m.Keeper.GetModuleAccountAddress() {
			if !gAcc.Amount.Equal(req.AllocatedAmount) {
				return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "allocated amount mismatch")
			}
			found = true
			break
		}
	}
	if !found {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "no genesis account for iro module account")
	}

	planId, err := m.Keeper.CreatePlan(ctx, req.AllocatedAmount, startTime, preLaunchTime, rollapp, req.BondingCurve, req.IncentivePlanParams, req.LiquidityPart, req.VestingDuration, req.VestingStartTimeAfterSettlement)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePlanResponse{
		PlanId: planId,
	}, nil
}

// CreatePlan creates a new IRO plan for a rollapp
// This function performs the following steps:
// 1. Sets the IRO plan to the rollapp with the specified pre-launch time.
// 2. Mints the allocated amount of tokens for the rollapp.
// 3. Creates a new plan with the provided parameters and validates it.
// 4. Creates a new module account for the IRO plan.
// 5. Charges the creation fee from the rollapp owner to the plan's module account.
// 6. Stores the plan in the keeper.
func (k Keeper) CreatePlan(ctx sdk.Context, allocatedAmount math.Int, start, preLaunchTime time.Time, rollapp rollapptypes.Rollapp, curve types.BondingCurve, incentivesParams types.IncentivePlanParams, liquidityPart math.LegacyDec, vestingDuration, vestingStartTimeAfterSettlement time.Duration) (string, error) {
	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Display, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	if err != nil {
		return "", err
	}

	plan := types.NewPlan(k.GetNextPlanIdAndIncrement(ctx), rollapp.RollappId, allocation, curve, start, preLaunchTime, incentivesParams, liquidityPart, vestingDuration, vestingStartTimeAfterSettlement)
	if err := plan.ValidateBasic(); err != nil {
		return "", errors.Join(gerrc.ErrInvalidArgument, err)
	}

	err = k.rk.SetIROPlanToRollapp(ctx, &rollapp, plan)
	if err != nil {
		return "", errors.Join(gerrc.ErrFailedPrecondition, err)
	}

	// Create a new module account for the IRO plan
	_, err = k.CreateModuleAccountForPlan(ctx, plan)
	if err != nil {
		return "", err
	}

	// charge creation fee
	feeAmt := k.GetParams(ctx).CreationFee
	cost := plan.BondingCurve.Cost(math.ZeroInt(), feeAmt)
	if !cost.IsPositive() {
		return "", errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid cost for fee charge")
	}

	feeCostInDym := sdk.NewCoin(appparams.BaseDenom, cost)
	err = k.BK.SendCoins(ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), plan.GetAddress(), sdk.NewCoins(feeCostInDym))
	if err != nil {
		return "", err
	}

	plan.SoldAmt = feeAmt
	plan.ClaimedAmt = feeAmt // set fee as claimed, as it's not claimable

	// Set the plan in the store
	k.SetPlan(ctx, plan)

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventNewIROPlan{
		Creator:   rollapp.Owner,
		PlanId:    fmt.Sprintf("%d", plan.Id),
		RollappId: rollapp.RollappId,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", plan.Id), nil
}

func (k Keeper) CreateModuleAccountForPlan(ctx sdk.Context, plan types.Plan) (sdk.ModuleAccountI, error) {
	moduleAccount := authtypes.NewEmptyModuleAccount(plan.ModuleAccName())
	moduleAccountI, ok := (k.AK.NewAccount(ctx, moduleAccount)).(sdk.ModuleAccountI)
	if !ok {
		return nil, errorsmod.Wrap(gerrc.ErrInternal, "failed to create module account")
	}
	k.AK.SetModuleAccount(ctx, moduleAccountI)
	return moduleAccountI, nil
}

// MintAllocation mints the allocated amount and registers the denom in the bank denom metadata store
func (k Keeper) MintAllocation(ctx sdk.Context, allocatedAmount math.Int, rollappId, rollappTokenSymbol string, exponent uint64) (sdk.Coin, error) {
	baseDenom := types.IRODenom(rollappId)
	displayDenom := types.IRODenom(rollappTokenSymbol)
	metadata := banktypes.Metadata{
		Description: fmt.Sprintf("IRO token for %s of rollapp %s", rollappTokenSymbol, rollappId),
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: baseDenom, Exponent: 0, Aliases: []string{}},
			{Denom: displayDenom, Exponent: uint32(exponent), Aliases: []string{}}, //nolint:gosec
		},
		Base:    baseDenom,
		Name:    displayDenom,
		Display: displayDenom,
		Symbol:  displayDenom,
	}
	if err := metadata.Validate(); err != nil {
		return sdk.Coin{}, errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), fmt.Sprintf("metadata: %v", metadata))
	}

	err := k.dk.CreateDenomMetadata(ctx, metadata)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "create denom metadata")
	}

	minted := sdk.NewCoin(baseDenom, allocatedAmount)
	err = k.BK.MintCoins(ctx, types.ModuleName, sdk.NewCoins(minted))
	if err != nil {
		return sdk.Coin{}, err
	}
	return minted, nil
}
