package transfergenesis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	uibc "github.com/dymensionxyz/dymension/v3/utils/ibc"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

/*
TODO: this whole file is temporary
	Prior to this we relied on the whitelist addr to set the canonical channel, but that is no longer possible
	This currently file is a hack (not secure)
	The real solution will come in a followup PR
	See https://github.com/dymensionxyz/research/issues/242
*/

type IBCModuleCanonicalChannelHack struct {
	porttypes.IBCModule // next one
	rollappKeeper       rollappkeeper.Keeper
	getRollapp          func(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool)
	channelKeeper       uibc.GetChannelClientState
}

func NewIBCModuleCanonicalChannelHack(
	next porttypes.IBCModule,
	getRollapp func(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool),
	channelKeeper uibc.GetChannelClientState,
) *IBCModuleCanonicalChannelHack {
	return &IBCModuleCanonicalChannelHack{IBCModule: next, getRollapp: getRollapp, channelKeeper: channelKeeper}
}

func (w IBCModuleCanonicalChannelHack) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := ctx.Logger().With("hack set canonical channel")

	chainID, err := uibc.ChainIDFromPortChannel(ctx, w.channelKeeper, packet.GetDestPort(), packet.GetDestChannel())
	if err == nil {
		ra, ok := w.getRollapp(ctx, chainID)
		if ok {
			ra.ChannelId = packet.GetDestChannel()
			w.rollappKeeper.SetRollapp(ctx, ra)

			l.Info("Set the canonical channel", "channel id", packet.GetDestChannel())
		}
	}
	return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
