package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	epochskeeper "github.com/osmosis-labs/osmosis/v15/x/epochs/keeper"

	streamerkeeper "github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	streamertypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// migrateStreamer creates epoch pointers for all epoch infos.
func migrateStreamer(ctx sdk.Context, sk streamerkeeper.Keeper, ek *epochskeeper.Keeper) error {
	for _, epoch := range ek.AllEpochInfos(ctx) {
		err := sk.SaveEpochPointer(ctx, streamertypes.NewEpochPointer(epoch.Identifier, epoch.Duration))
		if err != nil {
			return err
		}
	}
	return nil
}
