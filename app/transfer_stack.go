package app

import (
	packetforwardmiddleware "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/keeper"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v8"
	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcporttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee"
	delayedackmodule "github.com/dymensionxyz/dymension/v3/x/delayedack"
	denommetadatamodule "github.com/dymensionxyz/dymension/v3/x/denommetadata"
	ibccompletion "github.com/dymensionxyz/dymension/v3/x/ibc_completion"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/genesisbridge"
)

func (a *AppKeepers) InitTransferStack() {
	a.TransferStack = ibctransfer.NewIBCModule(a.TransferKeeper)

	a.TransferStack = ratelimit.NewIBCMiddleware(
		a.RateLimitingKeeper,
		a.TransferStack,
	)
	a.TransferStack = bridgingfee.NewIBCModule(
		a.TransferStack,
		*a.RollappKeeper,
		a.DelayedAckKeeper,
		a.TransferKeeper,
		*a.TxFeesKeeper,
	)
	a.TransferStack = packetforwardmiddleware.NewIBCMiddleware(
		a.TransferStack,
		a.PacketForwardMiddlewareKeeper,
		0,
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp,
	)
	a.TransferStack = ibccompletion.NewIBCModule(
		a.TransferStack,
		a.RollappKeeper,
		a.DelayedAckKeeper,
	)

	a.TransferStack = denommetadatamodule.NewIBCModule(a.TransferStack, a.DenomMetadataKeeper, a.RollappKeeper)
	// already instantiated in SetupHooks()
	a.DelayedAckMiddleware.Setup(
		delayedackmodule.WithIBCModule(a.TransferStack),
		delayedackmodule.WithKeeper(a.DelayedAckKeeper),
		delayedackmodule.WithRollappKeeper(a.RollappKeeper),
	)
	a.TransferStack = a.DelayedAckMiddleware
	a.TransferStack = genesisbridge.NewIBCModule(a.TransferStack, a.RollappKeeper, a.TransferKeeper, a.DenomMetadataKeeper)

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, a.TransferStack)
	a.IBCKeeper.SetRouter(ibcRouter)
}
