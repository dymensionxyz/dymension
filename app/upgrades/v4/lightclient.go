package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcchannelkeeper "github.com/cosmos/ibc-go/v7/modules/core/04-channel/keeper"

	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func migrateRollappLightClients(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper, lightClientKeeper lightclientkeeper.Keeper, ibcChannelKeeper ibcchannelkeeper.Keeper) {
	list := rollappkeeper.GetAllRollapps(ctx)
	for _, rollapp := range list {
		// check if the rollapp has a canonical channel already
		if rollapp.ChannelId == "" {
			return
		}
		// get the client ID the channel belongs to
		_, connection, err := ibcChannelKeeper.GetChannelConnection(ctx, ibctransfertypes.PortID, rollapp.ChannelId)
		if err != nil {
			// if could not find a connection, skip the canonical client assignment
			return
		}
		clientID := connection.GetClientID()
		// store the rollapp to canonical light client ID mapping
		lightClientKeeper.SetCanonicalClient(ctx, rollapp.RollappId, clientID)
	}
}
