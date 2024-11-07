package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

// migrateRollappGauges creates a gauge for each rollapp in the store
func migrateRollappGauges(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper, incentivizeKeeper *incentiveskeeper.Keeper) error {
	rollapps := rollappkeeper.GetAllRollapps(ctx)
	for _, rollapp := range rollapps {
		_, err := incentivizeKeeper.CreateRollappGauge(ctx, rollapp.RollappId)
		if err != nil {
			return err
		}
	}
	return nil
}
