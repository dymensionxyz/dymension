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

		// migrate bonding curve
		plan.BondingCurve.LiquidityDenomDecimals = 18
		plan.BondingCurve.RollappDenomDecimals = 18

		// nothing to set here
		plan.VestingPlan = types.IROVestingPlan{
			Amount:                   math.ZeroInt(),
			Claimed:                  math.ZeroInt(),
			VestingDuration:          0,
			StartTimeAfterSettlement: 0,
		}

		plan.LiquidityDenom = "adym"

		// liquidity part is 1.0 for old plans
		plan.LiquidityPart = math.LegacyOneDec()

		// max amount to sell is calculated from bonding curve
		eq := types.FindEquilibrium(plan.BondingCurve, plan.TotalAllocation.Amount, plan.LiquidityPart)
		plan.MaxAmountToSell = math.MaxInt(plan.SoldAmt, eq)

		plan.TradingEnabled = true

		plan.IroPlanDuration = plan.PreLaunchTime.Sub(plan.StartTime)

		if err := plan.ValidateBasic(); err != nil {
			panic(fmt.Errorf("invalid plan: %w", err))
		}

		// For settled plans, find and set the pool ID
		if plan.SettledDenom != "" {
			feeToken, err := m.k.tk.GetFeeToken(ctx, plan.SettledDenom)
			if err != nil {
				return fmt.Errorf("failed to get fee token for denom %s: %w", plan.SettledDenom, err)
			}
			if len(feeToken.Route) == 0 || feeToken.Route[0].TokenOutDenom != "adym" {
				return fmt.Errorf("fee token for denom %s does not have a route to adym", plan.SettledDenom)
			}
			plan.GraduatedPoolId = feeToken.Route[0].PoolId
		}

		m.k.SetPlan(ctx, plan)
	}

	return nil
}
