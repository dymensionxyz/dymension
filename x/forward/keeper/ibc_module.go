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

	/*
		SPAGHETTI CODE ALERT:

		*In general* we want a way to do something whenever an ibc transfer happens ("Hook"). It can happen
			1. on EIBC fulfill
			2. on finalize to the original recipient, for non fulfilled orders
			3. on finalize to the fulfiller, for fulfilled orders

		Chosen approach is a bit spaghetti:

		1. Do the hook on EIBC fulfillment, using immediate funds
		2. On finalize, look up the EIBC demand order to check if it's fulfilled or not.
			a. If it ISN'T, then do the hook AFTER the ibc transfer stack finishes
			b. If it IS, then do nothing

		We can do (2) by finding the eibc order directly using the packet key, because the status has not yet been update to finalized
	*/

	return m.IBCModule.OnRecvPacket(ctx, packet, relayer)
}
