package transfergenesis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	uibc "github.com/dymensionxyz/sdk-utils/utils/uibc"
)

/*
TODO: this whole file is temporary
	Prior to this we relied on the whitelist addr to set the canonical channel, but that is no longer possible
	This currently file is a hack (not secure)
	The real solution will come in a followup PR
	See https://github.com/dymensionxyz/research/issues/242
*/

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error) // implemented by ibc channel keeper
}

type IBCModuleCanonicalChannelHack struct {
	porttypes.IBCModule // next one
	rollappKeeper       rollappkeeper.Keeper
	channelKeeper       ChannelKeeper
}

func NewIBCModuleCanonicalChannelHack(
	next porttypes.IBCModule,
	rollappKeeper rollappkeeper.Keeper,
	channelKeeper uibc.ChainIDFromPortChannelKeeper,
) *IBCModuleCanonicalChannelHack {
	return &IBCModuleCanonicalChannelHack{IBCModule: next, rollappKeeper: rollappKeeper, channelKeeper: channelKeeper}
}

func (w IBCModuleCanonicalChannelHack) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	l := ctx.Logger().With("module", "hack set canonical channel")

	chainID, err := uibc.ChainIDFromPortChannel(ctx, w.channelKeeper, packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	ra, ok := w.rollappKeeper.GetRollapp(ctx, chainID)
	if ok && ra.ChannelId == "" {
		ra.ChannelId = packet.GetDestChannel()
		w.rollappKeeper.SetRollapp(ctx, ra)
		l.Info("Set the canonical channel.", "channel id", packet.GetDestChannel())
	}
	return w.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
