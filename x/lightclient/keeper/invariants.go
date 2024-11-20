package keeper

import (
	"errors"

	"cosmossdk.io/collections"
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{"client-state", InvariantClientState},
	{"attribution", InvariantAttribution},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantClientState(k Keeper) uinv.Func {
	return func(ctx sdk.Context) error {
		clients := k.GetAllCanonicalClients(ctx)
		var errs []error
		for _, client := range clients {
			errs = append(errs, checkClient(ctx, k, client))
		}
		// any error here is breaking
		return uinv.Breaking(errors.Join(errs...))
	}
}

func checkClient(ctx sdk.Context, k Keeper, client types.CanonicalClient) error {
	cs, ok := k.ibcClientKeeper.GetClientState(ctx, client.IbcClientId)
	if !ok {
		return gerrc.ErrNotFound.Wrapf("client state for client ID: %s", client.IbcClientId)
	}
	tmCS, ok := cs.(*ibctm.ClientState)
	if !ok {
		return gerrc.ErrInvalidArgument.Wrapf("client state is not a tendermint client state for client ID: %s", client.IbcClientId)
	}
	if tmCS.ChainId != client.RollappId {
		return gerrc.ErrInvalidArgument.Wrapf("client state chain ID does not match rollapp ID for client ID: %s: expect: %s", client.IbcClientId, client.RollappId)
	}
	_, ok = k.rollappKeeper.GetRollapp(ctx, client.RollappId)
	if !ok {
		return gerrc.ErrNotFound.Wrapf("rollapp for rollapp ID: %s", client.RollappId)
	}
	_, ok = k.ibcClientKeeper.GetClientConsensusState(ctx, client.IbcClientId, cs.GetLatestHeight())
	if !ok {
		return gerrc.ErrNotFound.Wrapf("latest consensus state for client ID: %s", client.IbcClientId)
	}
	return nil
}

func InvariantAttribution(k Keeper) uinv.Func {
	return func(ctx sdk.Context) error {

		err := k.headerSigners.Walk(ctx, nil, func(key collections.Triple[string, string, uint64]) (stop bool, err error) {
			seq := key.K1()
			clientID := key.K2()
			height := key.K3()

			_, err = k.clientHeightToSigner.Get(ctx, collections.Join(clientID, height))
			if err != nil {
				return false, errorsmod.Wrapf(err, "reverse lookup for sequencer address: %s", seq)
			}
			_, ok := k.ibcClientKeeper.GetClientConsensusState(ctx, clientID, clienttypes.NewHeight(1, height))
			if !ok {
				return false, gerrc.ErrNotFound.Wrapf("consensus state for client ID: %s", clientID)
			}
			signer, err := k.clientHeightToSigner.Get(ctx, collections.Join(clientID, height))
			if err != nil {
				return false, errorsmod.Wrapf(err, "get signer for client ID: %s", clientID)
			}
			if signer != seq {
				return false, gerrc.ErrInvalidArgument.Wrapf("signer mismatch: expected: %s, got: %s", seq, signer)
			}
			_, err = k.SeqK.RealSequencer(ctx, seq)
			if err != nil {
				return false, errorsmod.Wrapf(err, "get real sequencer for sequencer address: %s", seq)
			}
			return false, nil
		})

		if err != nil {
			return err
		}

		err = k.clientHeightToSigner.Walk(ctx, nil, func(key collections.Pair[string, uint64], seq string) (stop bool, err error) {
			clientID := key.K1()
			height := key.K2()
			ok, err := k.headerSigners.Has(ctx, collections.Join3(seq, clientID, height))
			if !ok || err != nil {
				return false, errorsmod.Wrapf(err, "forward lookup for client ID: %s", clientID)
			}
			return false, nil
		})

		return err, err != nil
	}
}
