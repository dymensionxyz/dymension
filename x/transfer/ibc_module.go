package transfer

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

type TransferHooks struct {
	Keeper eibckeeper.Keeper
}

// Should be called after packet finalization
// Recipient can either be the fulfiller of a hook that already occurred, or the original recipient still, who probably still wants the hook to happen
// NOTE: there is an asymmetry currently because on fulfill supports multiple hooks, but this finalization onRecv is hardcoded for x/forward atm
func (m *TransferHooks) AfterRecvPacket(ctx sdk.Context, p *commontypes.RollappPacket) {

	o, err := m.Keeper.PendingOrderByPacket(ctx, p)
	if errorsmod.IsOf(err, eibctypes.ErrDemandOrderDoesNotExist) {
		// not much we can do here, it should exist...
		return
	}
	if err != nil {
		// TODO: something
		return
	}

	if o.IsFulfilled() {
		// done
		return
	}

	if o.OnFulfillHook == nil {
		// done
		return
	}

	// TODO: do it

}
