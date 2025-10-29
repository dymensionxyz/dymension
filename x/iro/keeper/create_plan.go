package keeper

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

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

	rollapp, found := m.rk.GetRollapp(ctx, req.RollappId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "rollapp not found")
	}

	if rollapp.Owner != req.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	params := m.GetParams(ctx)

	// check minimal plan duration
	if req.IroPlanDuration < params.MinPlanDuration {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidDuration)
	}

	// check minimal liquidity part
	if req.LiquidityPart.LT(params.MinLiquidityPart) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "liquidity part must be at least %s", params.MinLiquidityPart)
	}

	// check vesting params
	if req.VestingDuration < params.MinVestingDuration {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "vesting duration must be at least %s", params.MinVestingDuration)
	}

	if req.VestingStartTimeAfterSettlement < params.MinVestingStartTimeAfterSettlement {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "vesting start time after settlement must be at least %s", params.MinVestingStartTimeAfterSettlement)
	}

	// validate incentive plan params
	if req.IncentivePlanParams.NumEpochsPaidOver < params.IncentivesMinNumEpochsPaidOver {
		return nil, errors.Join(gerrc.ErrInvalidArgument, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "num epochs paid over"))
	}
	if req.IncentivePlanParams.StartTimeAfterSettlement < params.IncentivesMinStartTimeAfterSettlement {
		return nil, errors.Join(gerrc.ErrInvalidArgument, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "start time after settlement"))
	}

	// Check if a plan already exists for the rollapp
	_, found = m.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrPlanExists)
	}

	found = false
	for _, gAcc := range rollapp.GenesisInfo.Accounts() {
		if gAcc.Address == m.GetModuleAccountAddress() {
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

	// validate rollapp decimals is correct
	if req.BondingCurve.RollappDenomDecimals != uint64(rollapp.GenesisInfo.NativeDenom.Exponent) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "rollapp decimals must be %d", rollapp.GenesisInfo.NativeDenom.Exponent)
	}

	// positive C is supported only for fixed price for now (due to equilibrium calculation)
	if !req.BondingCurve.C.IsZero() && (!req.BondingCurve.M.IsZero() && !req.BondingCurve.N.IsZero()) {
		return nil, errorsmod.Wrapf(types.ErrInvalidBondingCurve, "minimum price bonding curve is not supported")
	}

	// validate the liquidity denom is registered and curve decimals are correct
	liqToken, ok := m.BK.GetDenomMetaData(ctx, req.LiquidityDenom)
	if !ok {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "denom %s not registered", req.LiquidityDenom)
	}
	exponent := liqToken.DenomUnits[len(liqToken.DenomUnits)-1].Exponent
	if req.BondingCurve.LiquidityDenomDecimals != uint64(exponent) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "liquidity denom decimals must be %d", exponent)
	}

	// check liquidity denom is allowed
	if !slices.Contains(m.Keeper.gk.GetParams(ctx).AllowedPoolCreationDenoms, req.LiquidityDenom) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "denom not allowed")
	}

	planId, err := m.Keeper.CreatePlan(ctx,
		req.LiquidityDenom,
		req.AllocatedAmount,
		req.IroPlanDuration,
		req.StartTime,
		req.TradingEnabled,
		false,
		rollapp,
		req.BondingCurve,
		req.IncentivePlanParams,
		req.LiquidityPart,
		req.VestingDuration,
		req.VestingStartTimeAfterSettlement)
	if err != nil {
		return nil, err
	}

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventNewIROPlan{
		Creator:   rollapp.Owner,
		PlanId:    planId,
		RollappId: rollapp.RollappId,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreatePlanResponse{
		PlanId: planId,
	}, nil
}

// CreateStandardLaunchPlan creates a new IRO plan using global StandardLaunch parameters
// This function performs the following steps:
// 1. Validates the rollapp and owner authorization
// 2. Ensures 100% IRO allocation by comparing rollapp's InitialSupply with StandardLaunch allocation
// 3. Calculates M parameter for the bonding curve using converted target raise
// 4. Creates a plan with global StandardLaunch parameters and standard_launched = true
func (m msgServer) CreateStandardLaunchPlan(goCtx context.Context, req *types.MsgCreateStandardLaunchPlan) (*types.MsgCreatePlanResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	rollapp, found := m.rk.GetRollapp(ctx, req.RollappId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "rollapp not found")
	}

	if rollapp.Owner != req.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	params := m.GetParams(ctx)

	// Check if a plan already exists for the rollapp
	_, found = m.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrPlanExists)
	}

	// Validate 100% IRO allocation
	if !rollapp.GenesisInfo.InitialSupply.Equal(params.StandardLaunch.AllocationAmount) || len(rollapp.GenesisInfo.Accounts()) != 1 {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "rollapp must have 100%% IRO allocation: expected %s, got %s", params.StandardLaunch.AllocationAmount, rollapp.GenesisInfo.InitialSupply)
	}

	gAcc := rollapp.GenesisInfo.Accounts()[0]
	if gAcc.Address != m.GetModuleAccountAddress() || !gAcc.Amount.Equal(params.StandardLaunch.AllocationAmount) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "allocated amount mismatch")
	}

	// Validate liquidity denom is registered and allowed
	liqToken, ok := m.BK.GetDenomMetaData(ctx, req.LiquidityDenom)
	if !ok {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "denom %s not registered", req.LiquidityDenom)
	}
	liqTokenExponent := liqToken.DenomUnits[len(liqToken.DenomUnits)-1].Exponent

	// check liquidity denom is allowed
	if !slices.Contains(m.Keeper.gk.GetParams(ctx).AllowedPoolCreationDenoms, req.LiquidityDenom) {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "denom not allowed")
	}

	// Convert target raise from its original denom to the requested liquidity denom
	// This is needed because params.StandardLaunch.TargetRaise might be in a different denom
	convertedTargetRaise, err := m.convertTargetRaiseToLiquidityDenom(ctx, params.StandardLaunch.TargetRaise, req.LiquidityDenom)
	if err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "failed to convert target raise to liquidity denom: %v", err.Error())
	}

	// Calculate M parameter for the bonding curve
	// Convert amounts to decimal representation for calculation
	allocationDec := types.ScaleFromBase(params.StandardLaunch.AllocationAmount, int64(rollapp.GenesisInfo.NativeDenom.Exponent))
	evaluationDec := types.ScaleFromBase(convertedTargetRaise.Amount, int64(liqTokenExponent)).MulInt64(2)
	liquidityPart := math.LegacyOneDec()

	calculatedM := types.CalculateM(evaluationDec, allocationDec, params.StandardLaunch.CurveExp, liquidityPart)
	if !calculatedM.IsPositive() {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "calculated M parameter is not positive: %s", calculatedM)
	}

	// Create bonding curve with calculated M and global parameters
	bondingCurve := types.NewBondingCurve(
		calculatedM,
		params.StandardLaunch.CurveExp,
		math.LegacyZeroDec(),
		uint64(rollapp.GenesisInfo.NativeDenom.Exponent),
		uint64(liqTokenExponent),
	)

	// Validate the bonding curve
	if err := bondingCurve.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid bonding curve: %v", err.Error())
	}

	// Create plan using global StandardLaunch parameters
	planId, err := m.Keeper.CreatePlan(
		ctx,
		req.LiquidityDenom,
		params.StandardLaunch.AllocationAmount,
		0,           // no minimum plan duration
		time.Time{}, // no start time
		req.TradingEnabled,
		true, // standard launched
		rollapp,
		bondingCurve,
		types.IncentivePlanParams{}, // no incentive plan params for standard launch
		liquidityPart,
		0, // liquidity part for standard launch is 1.0, so no vesting duration
		0, // liquidity part for standard launch is 1.0, so no vesting start time after settlement
	)
	if err != nil {
		return nil, err
	}

	// Emit event
	err = uevent.EmitTypedEvent(ctx, &types.EventNewIROPlan{
		Creator:        rollapp.Owner,
		PlanId:         planId,
		RollappId:      rollapp.RollappId,
		StandardLaunch: true,
	})
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
func (k Keeper) CreatePlan(ctx sdk.Context, liquidityDenom string, allocatedAmount math.Int, planDuration time.Duration, startTime time.Time, tradingEnabled bool, standardLaunch bool, rollapp rollapptypes.Rollapp, curve types.BondingCurve, incentivesParams types.IncentivePlanParams, liquidityPart math.LegacyDec, vestingDuration, vestingStartTimeAfterSettlement time.Duration) (string, error) {
	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Display, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	if err != nil {
		return "", err
	}

	plan := types.NewPlan(k.GetNextPlanIdAndIncrement(ctx), rollapp.RollappId, liquidityDenom, allocation, curve, planDuration, incentivesParams, liquidityPart, vestingDuration, vestingStartTimeAfterSettlement)
	plan.StandardLaunch = standardLaunch

	// if trading enabled initially, set start time and pre-launch time
	if tradingEnabled {
		if startTime.Before(ctx.BlockTime()) {
			startTime = ctx.BlockTime()
		}
		plan.EnableTradingWithStartTime(startTime)
	}

	if err := plan.ValidateBasic(); err != nil {
		return "", errors.Join(gerrc.ErrInvalidArgument, err)
	}

	err = k.rk.SetIROPlanToRollapp(ctx, &rollapp, plan)
	if err != nil {
		return "", errors.Join(gerrc.ErrFailedPrecondition, err)
	}

	// set the fixed pre-launch time for custom and enabled plan
	// otherwise, prelaunch time will be updated once the requirement achieved
	if plan.TradingEnabled && !plan.StandardLaunch {
		k.rk.SetPreLaunchTime(ctx, &rollapp, plan.StartTime.Add(plan.IroPlanDuration))
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
		cost = math.NewInt(1) // charge minimum creation fee
	}

	feeCostLiquidlyCoin := sdk.NewCoin(plan.LiquidityDenom, cost)
	err = k.BK.SendCoins(ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), plan.GetAddress(), sdk.NewCoins(feeCostLiquidlyCoin))
	if err != nil {
		return "", err
	}

	plan.SoldAmt = feeAmt
	plan.ClaimedAmt = feeAmt // set fee as claimed, as it's not claimable

	// Set the plan in the store
	k.SetPlan(ctx, plan)

	return plan.GetID(), nil
}

// convertTargetRaiseToLiquidityDenom converts the target raise from its original denom to the requested liquidity denom
// If the denoms are the same, it returns the original target raise
// If they're different, it attempts to find a conversion path or returns an error
func (m msgServer) convertTargetRaiseToLiquidityDenom(ctx sdk.Context, targetRaise sdk.Coin, liquidityDenom string) (sdk.Coin, error) {
	// if the target raise denom is the same as the liquidity denom, return the original target raise
	if targetRaise.Denom == liquidityDenom {
		return targetRaise, nil
	}

	// convert the target raise to the base denom (just in case it's not set in base denom)
	baseTargetRaise, err := m.tk.CalcCoinInBaseDenom(ctx, targetRaise)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "failed to convert target raise to base denom: %v", err.Error())
	}

	// now get the target raise in the required liquidity denom
	liquidityTargetRaise, err := m.tk.CalcBaseInCoin(ctx, baseTargetRaise, liquidityDenom)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "failed to convert target raise to liquidity denom: %v", err.Error())
	}
	return liquidityTargetRaise, nil
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
