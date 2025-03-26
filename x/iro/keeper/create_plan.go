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

	rollapp, found := m.Keeper.rk.GetRollapp(ctx, req.RollappId)
	if !found {
		return nil, errorsmod.Wrapf(gerrc.ErrNotFound, "rollapp not found")
	}

	if rollapp.Owner != req.Owner {
		return nil, sdkerrors.ErrUnauthorized
	}

	params := m.Keeper.GetParams(ctx)

	// check minimal plan duration
	if req.IroPlanDuration < params.MinPlanDuration {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidEndTime)
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

	// validate rollapp decimals is correct
	if req.BondingCurve.RollappDenomDecimals != uint64(rollapp.GenesisInfo.NativeDenom.Exponent) {
		return nil, errorsmod.Wrapf(gerrc.ErrInvalidArgument, "rollapp decimals must be %d", rollapp.GenesisInfo.NativeDenom.Exponent)
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

	planId, err := m.Keeper.CreatePlan(ctx, req.LiquidityDenom, req.AllocatedAmount, req.IroPlanDuration, req.StartTime, req.TradingEnabled, rollapp, req.BondingCurve, req.IncentivePlanParams, req.LiquidityPart, req.VestingDuration, req.VestingStartTimeAfterSettlement)
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
func (k Keeper) CreatePlan(ctx sdk.Context, liquidityDenom string, allocatedAmount math.Int, planDuration time.Duration, startTime time.Time, tradingEnabled bool, rollapp rollapptypes.Rollapp, curve types.BondingCurve, incentivesParams types.IncentivePlanParams, liquidityPart math.LegacyDec, vestingDuration, vestingStartTimeAfterSettlement time.Duration) (string, error) {
	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Display, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	if err != nil {
		return "", err
	}

	plan := types.NewPlan(k.GetNextPlanIdAndIncrement(ctx), rollapp.RollappId, liquidityDenom, allocation, curve, planDuration, incentivesParams, liquidityPart, vestingDuration, vestingStartTimeAfterSettlement)

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

	feeCostLiquidlyCoin := sdk.NewCoin(plan.LiquidityDenom, cost)
	err = k.BK.SendCoins(ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), plan.GetAddress(), sdk.NewCoins(feeCostLiquidlyCoin))
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
