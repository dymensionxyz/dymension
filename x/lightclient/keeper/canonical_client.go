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
	k.Logger(ctx).Info("GetProspectiveCanonicalClient.")
	k.ibcClientKeeper.IterateClientStates(ctx, nil, func(client string, cs exported.ClientState) bool {
		k.Logger(ctx).Info("GetProspectiveCanonicalClient.", "client", client)
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

func (k Keeper) isValidClient(ctx sdk.Context, clientID string, cs exported.ClientState, rollappId string, maxHeight uint64) bool {
	l := ctx.Logger().With("callsite", "is valid client")
	tmClientState, ok := cs.(*ibctm.ClientState)
	if !ok {
		l.Info("not tm client") // TODO: remove
		return false
	}
	if tmClientState.ChainId != rollappId {
		l.Info("wrong chain id") // TODO: remove
		return false
	}
	if !types.IsCanonicalClientParamsValid(tmClientState) {
		l.Info("invalid params") // TODO: remove
		return false
	}
	res, err := k.ibcClientKeeper.ConsensusStateHeights(ctx, &ibcclienttypes.QueryConsensusStateHeightsRequest{
		ClientId:   clientID,
		Pagination: &query.PageRequest{Limit: maxHeight},
	})
	if err != nil {
		l.Info("failed to query cons state heights") // TODO: remove
		return false
	}
	l.Info("iterating cons state heights", "n", len(res.ConsensusStateHeights)) // TODO: remove
	atLeastOneMatch := false
	for _, consensusHeight := range res.ConsensusStateHeights {
		h := consensusHeight.GetRevisionHeight()
		l.Info("checking cons state", "height", h) // TODO: remove
		if maxHeight < h {
			l.Info("max height reached") // TODO: remove
			break
		}
		consensusState, _ := k.ibcClientKeeper.GetClientConsensusState(ctx, clientID, consensusHeight)
		tmConsensusState, _ := consensusState.(*ibctm.ConsensusState)
		stateInfoH, err := k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, h)
		if err != nil {
			l.Info("find state info for h") // TODO: remove
			return false
		}
		stateInfoHplus1, err := k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, h+1)
		if err != nil {
			l.Info("find state info for h+1") // TODO: remove
			return false
		}
		bd, _ := stateInfoH.GetBlockDescriptor(h)
		oldSequencer, err := k.GetSequencerPubKey(ctx, stateInfoHplus1.Sequencer)
		if err != nil {
			l.Info("get seq pub key") // TODO: remove
			return false
		}
		rollappState := types.RollappState{
			BlockDescriptor:    bd,
			NextBlockSequencer: oldSequencer,
		}
		err = types.CheckCompatibility(*tmConsensusState, rollappState)
		if err != nil {
			l.Info("incompat") // TODO: remove
			return false
		}
		atLeastOneMatch = true
	}
	// Need to be sure that at least one consensus state agrees with a state update
	// (There are also no disagreeing consensus states. There may be some consensus states
	// for future state updates, which will incur a fraud if they disagree.)
	return atLeastOneMatch
}
