package v4

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func migrateRollappRegisteredDenoms(ctx sdk.Context, rk *rollappkeeper.Keeper) error {
	// mainnet
	if _, found := rk.GetRollapp(ctx, nimRollappID); found {
		for _, denom := range nimDenoms {
			if err := rk.SetRegisteredDenom(ctx, nimRollappID, denom); err != nil {
				return fmt.Errorf("set registered denom: %s: %w", nimRollappID, err)
			}
		}
	}
	if _, found := rk.GetRollapp(ctx, mandeRollappID); found {
		for _, denom := range mandeDenoms {
			if err := rk.SetRegisteredDenom(ctx, mandeRollappID, denom); err != nil {
				return fmt.Errorf("set registered denom: %s: %w", mandeRollappID, err)
			}
		}
	}
	// testnet
	if _, found := rk.GetRollapp(ctx, rollappXRollappID); found {
		for _, denom := range rollappXDenoms {
			if err := rk.SetRegisteredDenom(ctx, rollappXRollappID, denom); err != nil {
				return fmt.Errorf("set registered denom: %s: %w", rollappXRollappID, err)
			}
		}
	}
	if _, found := rk.GetRollapp(ctx, crynuxRollappID); found {
		for _, denom := range crynuxDenoms {
			if err := rk.SetRegisteredDenom(ctx, crynuxRollappID, denom); err != nil {
				return fmt.Errorf("set registered denom: %s: %w", crynuxRollappID, err)
			}
		}
	}
	return nil
}
