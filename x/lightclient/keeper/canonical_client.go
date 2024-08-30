package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

// GetProspectiveCanonicalClient returns the client id of the first IBC client which can be set as the canonical client for the given rollapp.
// The canonical client criteria are:
// 1. The client must be a tendermint client.
// 2. The client state must match the expected client params as configured by the module
// 3. All the existing consensus states much match the corresponding height rollapp block descriptors
func (k Keeper) GetProspectiveCanonicalClient(ctx sdk.Context, rollappId string, maxHeight uint64) (clientID string, stateCompatible bool) {
	k.ibcClientKeeper.IterateClientStates(ctx, nil, func(client string, cs exported.ClientState) bool {
		ok := k.isValidClient(ctx, client, cs, rollappId, maxHeight)
		if ok {
			clientID = client
			stateCompatible = true
			return true
		}
		return false
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

func (k Keeper) isValidClient(ctx sdk.Context, clientID string, cs exported.ClientState, rollappId string, maxHeight uint64) bool {
	tmClientState, ok := cs.(*ibctm.ClientState)
	if !ok {
		return false
	}
	if tmClientState.ChainId != rollappId {
		return false
	}
	if !types.IsCanonicalClientParamsValid(tmClientState) {
		return false
	}
	res, err := k.ibcClientKeeper.ConsensusStateHeights(ctx, &ibcclienttypes.QueryConsensusStateHeightsRequest{
		ClientId:   clientID,
		Pagination: &query.PageRequest{Limit: maxHeight},
	})
	if err != nil {
		return false
	}
	for _, consensusHeight := range res.ConsensusStateHeights {
		h := consensusHeight.GetRevisionHeight()
		if maxHeight < h {
			break
		}
		consensusState, _ := k.ibcClientKeeper.GetClientConsensusState(ctx, clientID, consensusHeight)
		tmConsensusState, _ := consensusState.(*ibctm.ConsensusState)
		stateInfoH, err := k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, h)
		if err != nil {
			return false
		}
		stateInfoHplus1, err := k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, h+1)
		if err != nil {
			return false
		}
		bd, _ := stateInfoH.GetBlockDescriptor(h)
		oldSequencer, err := k.GetSequencerPubKey(ctx, stateInfoHplus1.Sequencer)
		if err != nil {
			return false
		}
		rollappState := types.RollappState{
			BlockDescriptor:    bd,
			NextBlockSequencer: oldSequencer,
		}
		err = types.CheckCompatibility(*tmConsensusState, rollappState)
		if err != nil {
			return false
		}
	}
	return true
}
