package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	k Keeper
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper Keeper) Migrator {
	return Migrator{k: keeper}
}

// Migrate1to2 migrates from version 1 to 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	// iterate over all plans and add missing fields
	store := prefix.NewStore(ctx.KVStore(m.k.storeKey), types.PlanKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var plan types.Plan
		m.k.cdc.MustUnmarshal(iterator.Value(), &plan)

		if err := plan.ValidateBasic(); err != nil {
			panic(fmt.Errorf("invalid plan: %v", err))
		}

		// migrate bonding curve
		plan.BondingCurve.LiquidityDenomDecimals = 18
		plan.BondingCurve.RollappDenomDecimals = 18

		// liquidity part is 1.0 for old plans
		plan.LiquidityPart = math.LegacyOneDec()

		// max amount to sell is calculated from bonding curve
		eq := types.FindEquilibrium(plan.BondingCurve, plan.TotalAllocation.Amount, plan.LiquidityPart)
		plan.MaxAmountToSell = math.MaxInt(plan.SoldAmt, eq)

		// nothing to set here
		plan.VestingPlan = types.IROVestingPlan{
			Amount:                   math.ZeroInt(),
			Claimed:                  math.ZeroInt(),
			VestingDuration:          0,
			StartTimeAfterSettlement: 0,
		}

		plan.TradingEnabled = true

		plan.IroPlanDuration = plan.PreLaunchTime.Sub(plan.StartTime)
		plan.LiquidityDenom = "adym"

		m.k.SetPlan(ctx, plan)
	}

	return nil
}
