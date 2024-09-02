package keeper

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
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
	// FIXME: get decimals from the caller / rollapp object
	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.Metadata.TokenSymbol, 18)
	if err != nil {
		return "", err
	}

	plan := types.NewPlan(k.GetLastPlanId(ctx)+1, rollapp.RollappId, allocation, curve, start, end)
	// Create a new module account for the IRO plan
	moduleAccountI, err := k.CreateModuleAccountForPlan(ctx, plan)
	if err != nil {
		return "", err
	}
	if plan.ModuleAccAddress != moduleAccountI.GetAddress().String() {
		return "", errorsmod.Wrap(gerrc.ErrInternal, "module account address mismatch")
	}

	// FIXME: charge creation fee

	// Set the plan in the store
	k.SetPlan(ctx, plan)
	k.SetLastPlanId(ctx, plan.Id)

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
func (k Keeper) MintAllocation(ctx sdk.Context, allocatedAmount math.Int, rollappId, rollappSymbolName string, exponent uint64) (sdk.Coin, error) {
	baseDenom := fmt.Sprintf("FUT_%s", rollappId)
	displayDenom := ""

	// FIXME: make the symbol name mandatory?
	if rollappSymbolName == "" {
		displayDenom = fmt.Sprintf("FUT_%s", strings.ToUpper(strings.TrimSuffix(rollappId, "_")))
	} else {
		displayDenom = fmt.Sprintf("FUT_%s", rollappSymbolName)
	}
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
