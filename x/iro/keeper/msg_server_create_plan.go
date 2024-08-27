package keeper

import (
	"fmt"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// MintAllocation mints the allocated amount and registers the denom in the bank denom metadata store
func (k Keeper) MintAllocation(ctx sdk.Context, allocatedAmount math.Int, rollappId, symbolName string) (sdk.Coin, error) {
	// Register the denom in the bank denom metadata store
	baseDenom := fmt.Sprintf("FUT_%s", rollappId)
	displayDenom := fmt.Sprintf("FUT_%s", symbolName)
	metadata := banktypes.Metadata{
		Description: fmt.Sprintf("Future token for rollapp %s", rollappId),
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: baseDenom, Exponent: 0, Aliases: []string{}},
			// FIXME: get decimals from API
			{Denom: displayDenom, Exponent: 18, Aliases: []string{}},
		},
		Base:    baseDenom,
		Name:    baseDenom,
		Display: symbolName,
		Symbol:  symbolName,
	}

	toMint := sdk.NewCoin(baseDenom, allocatedAmount)
	err := k.bk.MintCoins(ctx, types.ModuleName, sdk.NewCoins(toMint))
	if err != nil {
		return sdk.Coin{}, err
	}

	k.bk.SetDenomMetaData(ctx, metadata)
	return toMint, nil
}

func (k Keeper) CreatePlan(ctx sdk.Context, allocatedAmount math.Int, start, end time.Time, rollapp rollapptypes.Rollapp) (string, error) {
	if rollapp.GenesisChecksum == "" {
		return "", types.ErrRollappGenesisChecksumNotSet
	}

	if rollapp.Metadata.TokenSymbol == "" {
		return "", types.ErrRollappTokenSymbolNotSet
	}

	// rollapp cannot be sealed when creating a plan
	if rollapp.Sealed {
		return "", types.ErrRollappSealed
	}

	// Check if the plan already exists
	_, found := k.GetPlanByRollapp(ctx, rollapp.RollappId)
	if found {
		return "", types.ErrPlanExists
	}

	//validate end time is in the future
	if end.Before(ctx.BlockTime()) {
		return "", types.ErrInvalidEndTime
	}

	// FIXME: create a module account for the plan

	allocation, err := k.MintAllocation(ctx, allocatedAmount, rollapp.RollappId, rollapp.Metadata.TokenSymbol)
	if err != nil {
		return "", err
	}

	// FIXME: move the minted tokens to plan’s module account

	// Create a new plan
	plan := types.Plan{
		Id:               k.GetLastPlanId(ctx) + 1,
		RollappId:        rollapp.RollappId,
		ModuleAccAddress: "", // FIXME
		TotalAllocation:  allocation,
		Settled:          false,
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
