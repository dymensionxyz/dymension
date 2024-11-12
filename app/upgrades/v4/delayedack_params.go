package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func migrateDelayedAckParams(ctx sdk.Context, delayedAckKeeper delayedackkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := delayedacktypes.DefaultParams()

	// EpochIdentifier is the only one that hasn't changed
	params.EpochIdentifier = delayedAckKeeper.EpochIdentifier(ctx)

	delayedAckKeeper.SetParams(ctx, params)
}
