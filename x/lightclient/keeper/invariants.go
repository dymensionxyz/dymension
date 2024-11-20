package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/utils/invar"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

var invs = invar.NamedFuncsList[Keeper]{
	{"client-state", InvariantClientState},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantClientState(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		clients := k.GetAllCanonicalClients(ctx)
		for _, client := range clients {
			cs, ok := k.ibcClientKeeper.GetClientState(ctx, client.IbcClientId)
			if !ok {
				return gerrc.ErrNotFound.Wrapf("client state for client ID: %s", client.IbcClientId), true
			}
			tmCS, ok := cs.(*ibctm.ClientState)
			if !ok {
				return gerrc.ErrInvalidArgument.Wrapf("client state is not a tendermint client state for client ID: %s", client.IbcClientId), true
			}
			if tmCS.ChainId != client.RollappId {
				return gerrc.ErrInvalidArgument.Wrapf("client state chain ID does not match rollapp ID for client ID: %s: expect: %s", client.IbcClientId, client.RollappId), true
			}
			_, ok = k.rollappKeeper.GetRollapp(ctx, client.RollappId)
			if !ok {
				return gerrc.ErrNotFound.Wrapf("rollapp for rollapp ID: %s", client.RollappId), true
			}
			_, ok = k.ibcClientKeeper.GetClientConsensusState(ctx, client.IbcClientId, cs.GetLatestHeight())
			if !ok {
				return gerrc.ErrNotFound.Wrapf("latest consensus state for client ID: %s", client.IbcClientId), true
			}
		}
		return nil, false
	}
}
