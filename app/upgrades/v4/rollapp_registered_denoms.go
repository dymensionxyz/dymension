package v4

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func migrateRollappRegisteredDenoms(ctx sdk.Context, rk *rollappkeeper.Keeper) error {
	for _, denom := range nimDenoms {
		if err := rk.SetRegisteredDenom(ctx, nimRollappID, denom); err != nil {
			return fmt.Errorf("set registered denom: %s: %w", nimRollappID, err)
		}
	}
	for _, denom := range mandeDenoms {
		if err := rk.SetRegisteredDenom(ctx, mandeRollappID, denom); err != nil {
			return fmt.Errorf("set registered denom: %s: %w", mandeRollappID, err)
		}
	}
	return nil
}
