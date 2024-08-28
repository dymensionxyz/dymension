package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	ibcRevisionNumber = 1
)

// GetProspectiveCanonicalClient returns the client id of the first IBC client which can be set as the canonical client for the given rollapp.
// The canonical client criteria are:
// 1. The client must be a tendermint client.
// 2. The client state must match the expected client params as configured by the module
// 3. All the block descriptors in the state info must be compatible with the client consensus state
func (k Keeper) GetProspectiveCanonicalClient(ctx sdk.Context, rollappId string, stateInfo *rollapptypes.StateInfo) (clientID string, foundClient bool) {
	clients := k.ibcClientKeeper.GetAllGenesisClients(ctx)
	for _, client := range clients {
		clientState, err := ibcclienttypes.UnpackClientState(client.ClientState)
		if err != nil {
			continue
		}
		// Cast client state to tendermint client state - we need this to get the chain id and state height
		tmClientState, ok := clientState.(*ibctm.ClientState)
		if !ok {
			continue
		}
		if tmClientState.ChainId != rollappId {
			continue
		}
		if !types.IsCanonicalClientParamsValid(tmClientState) {
			continue
		}
		sequencerPk, err := k.GetSequencerPubKey(ctx, stateInfo.Sequencer)
		if err != nil {
			continue
		}
		for _, bd := range stateInfo.GetBDs().BD {
			height := ibcclienttypes.NewHeight(ibcRevisionNumber, bd.GetHeight())
			consensusState, found := k.ibcClientKeeper.GetClientConsensusState(ctx, client.ClientId, height)
			if !found {
				continue
			}
			// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
			tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
			if !ok {
				continue
			}
			ibcState := types.IBCState{
				Root:               tmConsensusState.GetRoot().GetHash(),
				NextValidatorsHash: tmConsensusState.NextValidatorsHash,
				Timestamp:          time.Unix(0, int64(tmConsensusState.GetTimestamp())),
			}
			rollappState := types.RollappState{
				BlockDescriptor: bd,
			}
			// Check if BD for next block exists in same stateinfo
			if stateInfo.ContainsHeight(bd.GetHeight() + 1) {
				rollappState.NextBlockSequencer = sequencerPk
			}
			err := types.CheckCompatibility(ibcState, rollappState)
			if err != nil {
				continue
			}
			clientID = client.GetClientId()
			foundClient = true
			return
		}
	}
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
