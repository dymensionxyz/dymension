package keeper

// DONTCOVER

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, keeper Keeper) {
	ir.RegisterRoute(types.ModuleName, "accumulation-store-invariant", AccumulationStoreInvariant(keeper))
	ir.RegisterRoute(types.ModuleName, "locks-amount-invariant", LocksBalancesInvariant(keeper))
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		for _, inv := range []sdk.Invariant{
			AccumulationStoreInvariant(k),
			LocksBalancesInvariant(k),
		} {
			res, stop := inv(ctx)
			if stop {
				return res, stop
			}
		}
		return "", false
	}
}

// AccumulationStoreInvariant ensures that the sum of all lockups at a given duration
// is equal to the value stored within the accumulation store.
func AccumulationStoreInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		moduleAcc := keeper.ak.GetModuleAccount(ctx, types.ModuleName)
		balances := keeper.bk.GetAllBalances(ctx, moduleAcc.GetAddress())

		// check 1s, 1 day, 1 week, 2 weeks
		durations := []time.Duration{
			time.Second,
			time.Hour * 24,
			time.Hour * 24 * 7,
			time.Hour * 24 * 14,
		}

		// loop all denoms on lockup module
		for _, coin := range balances {
			denom := coin.Denom
			for _, duration := range durations {
				accumulation := keeper.GetPeriodLocksAccumulation(ctx, types.QueryCondition{
					LockQueryType: types.ByDuration,
					Denom:         denom,
					Duration:      duration,
				})

				locks := keeper.GetLocksLongerThanDurationDenom(ctx, denom, duration)
				lockupSum := sdk.ZeroInt()
				for _, lock := range locks {
					lockupSum = lockupSum.Add(lock.Coins.AmountOf(denom))
				}

				if !accumulation.Equal(lockupSum) {
					return sdk.FormatInvariant(types.ModuleName, "accumulation-store-invariant",
						fmt.Sprintf("\taccumulation store value does not fit actual lockup sum: %s != %s\n",
							accumulation.String(), lockupSum.String(),
						)), true
				}
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "accumulation-store-invariant", "All lockup accumulation invariant passed"), false
	}
}

// LocksBalancesInvariant ensure that the module balance and the sum of all
// tokens within all locks have the equivalent amount of tokens.
func LocksBalancesInvariant(keeper Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		moduleAcc := keeper.ak.GetModuleAccount(ctx, types.ModuleName)
		balances := keeper.bk.GetAllBalances(ctx, moduleAcc.GetAddress())

		// loop all denoms on lockup module
		for _, coin := range balances {
			denom := coin.Denom
			lockedAmount := sdk.ZeroInt()
			locksByDenom := keeper.GetLocksDenom(ctx, denom)
			for _, lock := range locksByDenom {
				lockedAmount = lockedAmount.Add(lock.Coins.AmountOf(denom))
			}
			if !lockedAmount.Equal(coin.Amount) {
				return sdk.FormatInvariant(types.ModuleName, "locks-amount-invariant",
					fmt.Sprintf("\tlocks amount of %s does not fit actual module balance: %s != %s\n",
						denom, lockedAmount.String(), coin.Amount.String(),
					)), true
			}
		}

		return sdk.FormatInvariant(types.ModuleName, "locks-amount-invariant", "All lockup amount invariant passed"), false
	}
}
