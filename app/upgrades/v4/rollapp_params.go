package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func migrateRollappParams(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := rollapptypes.DefaultParams()

	// Dispute period is the only one that hasn't changed
	params.DisputePeriodInBlocks = rollappkeeper.DisputePeriodInBlocks(ctx)

	rollappkeeper.SetParams(ctx, params)
}
