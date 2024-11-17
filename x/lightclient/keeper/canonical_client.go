package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// GetProspectiveCanonicalClient returns the client id of the first IBC client which can be set as the canonical client for the given rollapp.
// The canonical client criteria are:
// 1. The client must be a tendermint client.
// 2. The client state must match the expected client params as configured by the module
// 3. The client state must have the same chain id as the rollapp id.
// 4. All the existing consensus states much match the corresponding height rollapp block descriptors
func (k Keeper) GetProspectiveCanonicalClient(ctx sdk.Context, sInfo *rollapptypes.StateInfo) (clientID string, found bool) {
	k.ibcClientKeeper.IterateClientStates(ctx, nil, func(client string, cs exported.ClientState) bool {
		rollappId := sInfo.GetRollappId()

		tmClientState, ok := cs.(*ibctm.ClientState)
		if !ok {
			return false
		}
		if tmClientState.ChainId != rollappId {
			return false
		}

		expClient := k.expectedClient()
		if err := types.IsCanonicalClientParamsValid(tmClientState, &expClient); err != nil {
			k.Logger(ctx).Debug("validate client params against light client with same chain id",
				"rollapp", sInfo.GetRollappId(),
				"client", client,
				"expected", expClient,
				"actual", tmClientState,
				"err", err,
			)
			return false
		}

		// validate state info against optimistically accepted headers
		validated, err := k.ValidateOptimisticUpdates(ctx, client, sInfo)
		if err != nil {
			k.Logger(ctx).Debug("validate rollapp state against light client with same chain id",
				"rollapp", sInfo.GetRollappId(),
				"client", client,
				"err", err,
			)
			return false
		}
		if !validated {
			return false
		}

		// we successfully validated the state info against a potential client
		clientID = client
		found = true
		return true
	})
	return
}

func (k Keeper) GetCanonicalClient(ctx sdk.Context, rollappId string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRollappClientKey(rollappId))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetCanonicalClient(ctx sdk.Context, rollappId string, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRollappClientKey(rollappId), []byte(clientID))
	store.Set(types.CanonicalClientKey(clientID), []byte(rollappId))
	// TODO: event and log
}

func (k Keeper) GetAllCanonicalClients(ctx sdk.Context) (clients []types.CanonicalClient) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RollappClientKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		clients = append(clients, types.CanonicalClient{
			RollappId:   string(iterator.Key()[1:]),
			IbcClientId: string(iterator.Value()),
		})
	}
	return
}

func (k Keeper) expectedClient() ibctm.ClientState {
	return types.DefaultExpectedCanonicalClientParams()
}

func (k Keeper) ValidateHeaderAgainstStateInfo(ctx sdk.Context, sInfo *rollapptypes.StateInfo, consState *ibctm.ConsensusState, h uint64) error {
	bd, ok := sInfo.GetBlockDescriptor(h)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInternal, "no block descriptor found for height %d", h)
	}
	nextSeq, err := k.SeqK.RealSequencer(ctx, sInfo.NextSequencerForHeight(h))
	if err != nil {
		return errorsmod.Wrap(errors.Join(err, gerrc.ErrInternal), "get sequencer of state info")
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: nextSeq,
	}
	return errorsmod.Wrap(types.CheckCompatibility(*consState, rollappState), "check compatibility")
}
