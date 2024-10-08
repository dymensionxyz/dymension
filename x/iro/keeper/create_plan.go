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
	// check minimal duration
	if req.IroPlanDuration < m.Keeper.GetParams(ctx).MinPlanDuration {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrInvalidEndTime)
	}
	preLaunchTime := startTime.Add(req.IroPlanDuration)

	// validate incentive plan params
	params := m.Keeper.GetParams(ctx)
	if req.IncentivePlanParams.NumEpochsPaidOver < params.IncentivesMinNumEpochsPaidOver {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "num epochs paid over"))
	}
	if req.IncentivePlanParams.StartTimeAfterSettlement < params.IncentivesMinStartTimeAfterSettlement {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, errorsmod.Wrap(types.ErrInvalidIncentivePlanParams, "start time after settlement"))
	}

	// Check if the plan already exists
	_, found = m.Keeper.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return nil, errors.Join(gerrc.ErrFailedPrecondition, types.ErrPlanExists)
	}

	// validate the IRO funding genesis transfer is expected in rollapp's genesis info
	if rollapp.GenesisInfo.GenesisAccounts == nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis accounts not found")
	}
	found = false
	for _, gAcc := range rollapp.GenesisInfo.GenesisAccounts.Accounts {
		if gAcc.Address == m.Keeper.GetModuleAccountAddress() {
			if !gAcc.Amount.Equal(req.AllocatedAmount) {
				return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "allocated amount mismatch")
			}
			found = true
			break
		}
	}
	if !found {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer not found for IRO")
	}

	planId, err := m.Keeper.CreatePlan(ctx, req.AllocatedAmount, startTime, preLaunchTime, rollapp, req.BondingCurve, req.IncentivePlanParams)
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
func (k Keeper) CreatePlan(ctx sdk.Context, allocatedAmount math.Int, start, preLaunchTime time.Time, rollapp rollapptypes.Rollapp, curve types.BondingCurve, incentivesParams types.IncentivePlanParams) (string, error) {
	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.GenesisInfo.NativeDenom.Display, uint64(rollapp.GenesisInfo.NativeDenom.Exponent))
	if err != nil {
		return "", err
	}
	plan := types.NewPlan(k.GetNextPlanIdAndIncrement(ctx), rollapp.RollappId, allocation, curve, start, preLaunchTime, incentivesParams)
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
	fee := sdk.NewCoin(appparams.BaseDenom, k.GetParams(ctx).CreationFee)
	err = k.BK.SendCoins(ctx, sdk.MustAccAddressFromBech32(rollapp.Owner), plan.GetAddress(), sdk.NewCoins(fee))
	if err != nil {
		return "", err
	}

	// Set the plan in the store
	k.SetPlan(ctx, plan)

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
