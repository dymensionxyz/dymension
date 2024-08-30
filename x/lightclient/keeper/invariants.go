package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

// RegisterInvariants registers the lightclient module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "canonical-client-valid", CanonicalClientsValid(k))
}

// CanonicalClientsValid checks that all canonical clients have a known rollapp as their chain ID
func CanonicalClientsValid(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			broken bool
			msg    string
		)
		clients := k.GetAllCanonicalClients(ctx)
		for _, client := range clients {
			cs, found := k.ibcClientKeeper.GetClientState(ctx, client.IbcClientId)
			if !found {
				broken = true
				msg += "client state not found for client ID " + client.IbcClientId + "\n"
			}
			tmCS, ok := cs.(*ibctm.ClientState)
			if !ok {
				broken = true
				msg += "client state is not a tendermint client state for client ID " + client.IbcClientId + "\n"
			}
			if tmCS.ChainId != client.RollappId {
				broken = true
				msg += "client state chain ID does not match rollapp ID for client " + client.IbcClientId + "\n"
			}
			_, found = k.rollappKeeper.GetRollapp(ctx, client.RollappId)
			if !found {
				broken = true
				msg += "rollapp not found for given rollapp ID " + client.RollappId + "\n"
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, "canonical-client-valid",
			msg,
		), broken
	}
}
