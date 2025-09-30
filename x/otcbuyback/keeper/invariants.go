// x/otcbuyback/keeper/invariants.go
package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

// RegisterInvariants registers all otcbuyback invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "module-account-balance", ModuleAccountBalanceInvariant(k))
}

// ModuleAccountBalanceInvariant checks that module account has sufficient tokens
// to cover all outstanding obligations
func ModuleAccountBalanceInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// Calculate total outstanding obligations
		expecetedBalance := sdk.Coins{}

		auctions, err := k.GetAllAuctions(ctx, false)
		if err != nil {
			return sdk.FormatInvariant(types.ModuleName, "module-account-balance",
				fmt.Sprintf("failed to get auctions: %v", err)), true
		}

		for _, auction := range auctions {
			// basic validation
			if auction.SoldAmount.GT(auction.Allocation) {
				return sdk.FormatInvariant(types.ModuleName, "module-account-balance",
					fmt.Sprintf("auction %d: sold amount %s > allocation %s",
						auction.Id, auction.SoldAmount, auction.Allocation)), true
			}

			totalPurchased := math.ZeroInt()
			unclaimedAmount := math.ZeroInt()
			rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](auction.Id)
			err := k.purchases.Walk(ctx, rng, func(key collections.Pair[uint64, sdk.AccAddress], purchase types.Purchase) (bool, error) {
				// Check: Claimed >= 0
				if purchase.Claimed.IsNegative() {
					return true, fmt.Errorf("auction %d, buyer %s: negative claimed amount %s",
						auction.Id, key.K2(), purchase.Claimed)
				}

				// Check: Claimed <= Amount
				if purchase.Claimed.GT(purchase.Amount) {
					return true, fmt.Errorf("auction %d, buyer %s: claimed %s exceeds purchased %s",
						auction.Id, key.K2(), purchase.Claimed, purchase.Amount)
				}

				// add unclaimed amount to unclaimed amount
				unclaimedAmount = unclaimedAmount.Add(purchase.UnclaimedAmount())

				// add to total purchased
				totalPurchased = totalPurchased.Add(purchase.Amount)

				return false, nil
			})
			if err != nil {
				return sdk.FormatInvariant(types.ModuleName, "module-account-balance",
					fmt.Sprintf("failed to walk purchases: %v", err)), true
			}

			// assert that total purchased is equal to the sold amount
			if !totalPurchased.Equal(auction.SoldAmount) {
				return sdk.FormatInvariant(types.ModuleName, "module-account-balance",
					fmt.Sprintf("total purchased %s != sold amount %s",
						totalPurchased, auction.SoldAmount)), true
			}

			expecetedBalance = expecetedBalance.Add(sdk.NewCoin(k.baseDenom, unclaimedAmount))

			// For active auctions: remaining allocation and raised amount is expected to be in the module account
			if !auction.IsCompleted() {
				expecetedBalance = expecetedBalance.Add(sdk.NewCoin(k.baseDenom, auction.GetRemainingAllocation()))
				expecetedBalance = expecetedBalance.Add(auction.GetRaisedAmount()...)
			}
		}

		// Check module account has sufficient base tokens
		moduleBalance := k.bankKeeper.GetAllBalances(ctx, k.GetModuleAccountAddress())

		if expecetedBalance.IsAnyGT(moduleBalance) {
			return sdk.FormatInvariant(types.ModuleName, "module-account-balance",
				fmt.Sprintf("insufficient module balance: have %s, need %s",
					moduleBalance, expecetedBalance)), true
		}

		return sdk.FormatInvariant(types.ModuleName, "module-account-balance", "module account balance is sufficient"), false
	}
}
