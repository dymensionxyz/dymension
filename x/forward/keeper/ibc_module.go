package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	exported "github.com/cosmos/ibc-go/v8/modules/core/exported"
)

type IBCModule struct {
	porttypes.IBCModule
	Keeper
}

func NewIBCModule(k Keeper, next porttypes.IBCModule) *IBCModule {
	return &IBCModule{Keeper: k, IBCModule: next}
}

// Should be called after packet finalization
// Recipient can either be the fulfiller of a hook that already occurred, or the original recipient still, who probably still wants the hook to happen
// NOTE: there is an asymmetry currently because on fulfill supports multiple hooks, but this finalization onRecv is hardcoded for x/forward atm
func (m *IBCModule) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
