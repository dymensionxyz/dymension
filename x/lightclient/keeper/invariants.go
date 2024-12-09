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
	{Name: "client-state", Func: InvariantClientState},
	{Name: "attribution", Func: InvariantAttribution},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

// client state should match rollapp and have a consensus state
func InvariantClientState(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		clients := k.GetAllCanonicalClients(ctx)
		var errs []error
		for _, client := range clients {
			errs = append(errs, checkClient(ctx, k, client))
		}
		return errors.Join(errs...)
	})
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
	// NOTE: it is not actually an invariant that the consensus state needs to exist for the latest height
	// Because they can be pruned
	return nil
}

// the indexes used to attribute fraud should be populated and properly pruned
func InvariantAttribution(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		err := k.headerSigners.Walk(ctx, nil, func(key collections.Triple[string, string, uint64]) (bool, error) {
			seq := key.K1()
			clientID := key.K2()
			height := key.K3()
			err := checkAttributionIndexes(k, ctx, clientID, height, seq)
			err = errorsmod.Wrapf(err, "header signer for sequencer: %s, client ID: %s, height: %d", seq, clientID, height)
			errs = append(errs, err)
			return false, nil
		})
		errs = append(errs, err)
		if err := errors.Join(errs...); err != nil {
			return errorsmod.Wrap(err, "check header signers")
		}

		errs = nil

		err = k.clientHeightToSigner.Walk(ctx, nil, func(key collections.Pair[string, uint64], seq string) (bool, error) {
			clientID := key.K1()
			height := key.K2()
			ok, err := k.headerSigners.Has(ctx, collections.Join3(seq, clientID, height))
			if !ok || err != nil {
				errs = append(errs, gerrc.ErrNotFound.Wrapf("header signer for sequencer: %s, client ID: %s, height: %d", seq, clientID, height))
			}
			return false, nil
		})
		errs = append(errs, err)
		return errors.Join(errs...)
	})
}

func checkAttributionIndexes(k Keeper, ctx sdk.Context, clientID string, height uint64, seq string) error {
	_, err := k.clientHeightToSigner.Get(ctx, collections.Join(clientID, height))
	if err != nil {
		return errorsmod.Wrapf(err, "reverse lookup for sequencer address: %s", seq)
	}
	_, ok := k.ibcClientKeeper.GetClientConsensusState(ctx, clientID, clienttypes.NewHeight(1, height))
	if !ok {
		return gerrc.ErrNotFound.Wrapf("consensus state for client ID: %s", clientID)
	}
	signer, err := k.clientHeightToSigner.Get(ctx, collections.Join(clientID, height))
	if err != nil {
		return errorsmod.Wrapf(err, "get signer for client ID: %s", clientID)
	}
	if signer != seq {
		return gerrc.ErrInvalidArgument.Wrapf("signer mismatch: expected: %s, got: %s", seq, signer)
	}
	_, err = k.SeqK.RealSequencer(ctx, seq)
	if err != nil {
		return errorsmod.Wrapf(err, "get real sequencer for sequencer address: %s", seq)
	}
	return nil
}
