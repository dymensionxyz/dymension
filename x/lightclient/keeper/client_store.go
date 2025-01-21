package keeper

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
)

// IterateConsensusStateDescending iterates through all consensus states in descending order
// until cb returns true.
func IterateConsensusStateDescending(clientStore storetypes.KVStore, cb func(height exported.Height) (stop bool)) {
	iterator := storetypes.KVStoreReversePrefixIterator(clientStore, []byte(ibctm.KeyIterateConsensusStatePrefix))
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		iterKey := iterator.Key()
		height := ibctm.GetHeightFromIterationKey(iterKey)
		if cb(height) {
			break
		}
	}
}

// functions here copied from ibc-go/modules/core/02-client/keeper/
// as we need direct access to the client store

// getClientState returns the client state for a particular client
func getClientState(clientStore storetypes.KVStore, cdc codec.BinaryCodec) exported.ClientState {
	bz := clientStore.Get(host.ClientStateKey())
	if len(bz) == 0 {
		return nil
	}

	return clienttypes.MustUnmarshalClientState(cdc, bz)
}

// must be tendermint!
func getClientStateTM(clientStore storetypes.KVStore, cdc codec.BinaryCodec) *ibctm.ClientState {
	c := getClientState(clientStore, cdc)
	tmClientState, _ := c.(*ibctm.ClientState)
	return tmClientState
}

// setClientState stores the client state
func setClientState(clientStore storetypes.KVStore, cdc codec.BinaryCodec, clientState exported.ClientState) {
	key := host.ClientStateKey()
	val := clienttypes.MustMarshalClientState(cdc, clientState)
	clientStore.Set(key, val)
}

func setConsensusState(clientStore storetypes.KVStore, cdc codec.BinaryCodec, height exported.Height, cs exported.ConsensusState) {
	key := host.ConsensusStateKey(height)
	val := clienttypes.MustMarshalConsensusState(cdc, cs)
	clientStore.Set(key, val)
}

// setConsensusMetadata sets context time as processed time and set context height as processed height
// as this is internal tendermint light client logic.
// client state and consensus state will be set by client keeper
// set iteration key to provide ability for efficient ordered iteration of consensus states.
func setConsensusMetadata(ctx sdk.Context, clientStore storetypes.KVStore, height exported.Height) {
	setConsensusMetadataWithValues(clientStore, height, clienttypes.GetSelfHeight(ctx), uint64(ctx.BlockTime().UnixNano()))
}

// setConsensusMetadataWithValues sets the consensus metadata with the provided values
func setConsensusMetadataWithValues(
	clientStore storetypes.KVStore, height,
	processedHeight exported.Height,
	processedTime uint64,
) {
	ibctm.SetProcessedTime(clientStore, height, processedTime)
	ibctm.SetProcessedHeight(clientStore, height, processedHeight)
	ibctm.SetIterationKey(clientStore, height)
}

// deleteConsensusMetadata deletes the metadata stored for a particular consensus state.
func deleteConsensusMetadata(clientStore storetypes.KVStore, height exported.Height) {
	deleteProcessedTime(clientStore, height)
	deleteProcessedHeight(clientStore, height)
	deleteIterationKey(clientStore, height)
}

// deleteConsensusState deletes the consensus state at the given height
func deleteConsensusState(clientStore storetypes.KVStore, height exported.Height) {
	key := host.ConsensusStateKey(height)
	clientStore.Delete(key)
}

// deleteProcessedTime deletes the processedTime for a given height
func deleteProcessedTime(clientStore storetypes.KVStore, height exported.Height) {
	key := ibctm.ProcessedTimeKey(height)
	clientStore.Delete(key)
}

// deleteProcessedHeight deletes the processedHeight for a given height
func deleteProcessedHeight(clientStore storetypes.KVStore, height exported.Height) {
	key := ibctm.ProcessedHeightKey(height)
	clientStore.Delete(key)
}

// deleteIterationKey deletes the iteration key for a given height
func deleteIterationKey(clientStore storetypes.KVStore, height exported.Height) {
	key := ibctm.IterationKey(height)
	clientStore.Delete(key)
}

// GetFirstHeight returns the lowest height available for a client.
func (k Keeper) GetFirstConsensusStateHeight(ctx sdk.Context, clientID string) uint64 {
	height := clienttypes.Height{}
	k.ibcClientKeeper.IterateConsensusStates(ctx, func(clientID string, cs clienttypes.ConsensusStateWithHeight) bool {
		height = cs.Height
		return true
	})
	return height.GetRevisionHeight()
}
