package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

// functions here copied from ibc-go/modules/core/02-client/keeper/
// as we need direct access to the client store

// getClientState returns the client state for a particular client
func getClientState(clientStore sdk.KVStore, cdc codec.BinaryCodec) exported.ClientState {
	bz := clientStore.Get(host.ClientStateKey())
	if len(bz) == 0 {
		return nil
	}

	return clienttypes.MustUnmarshalClientState(cdc, bz)
}

// setClientState stores the client state
func setClientState(clientStore sdk.KVStore, cdc codec.BinaryCodec, clientState exported.ClientState) {
	key := host.ClientStateKey()
	val := clienttypes.MustMarshalClientState(cdc, clientState)
	clientStore.Set(key, val)
}

func setConsensusState(clientStore sdk.KVStore, cdc codec.BinaryCodec, height exported.Height, cs exported.ConsensusState) {
	key := host.ConsensusStateKey(height)
	val := clienttypes.MustMarshalConsensusState(cdc, cs)
	clientStore.Set(key, val)
}

// setConsensusMetadata sets context time as processed time and set context height as processed height
// as this is internal tendermint light client logic.
// client state and consensus state will be set by client keeper
// set iteration key to provide ability for efficient ordered iteration of consensus states.
func setConsensusMetadata(ctx sdk.Context, clientStore sdk.KVStore, height exported.Height) {
	setConsensusMetadataWithValues(clientStore, height, clienttypes.GetSelfHeight(ctx), uint64(ctx.BlockTime().UnixNano()))
}

// setConsensusMetadataWithValues sets the consensus metadata with the provided values
func setConsensusMetadataWithValues(
	clientStore sdk.KVStore, height,
	processedHeight exported.Height,
	processedTime uint64,
) {
	ibctm.SetProcessedTime(clientStore, height, processedTime)
	ibctm.SetProcessedHeight(clientStore, height, processedHeight)
	ibctm.SetIterationKey(clientStore, height)
}

// deleteConsensusMetadata deletes the metadata stored for a particular consensus state.
func deleteConsensusMetadata(clientStore sdk.KVStore, height exported.Height) {
	deleteProcessedTime(clientStore, height)
	deleteProcessedHeight(clientStore, height)
	deleteIterationKey(clientStore, height)
}

// deleteConsensusState deletes the consensus state at the given height
func deleteConsensusState(clientStore sdk.KVStore, height exported.Height) {
	key := host.ConsensusStateKey(height)
	clientStore.Delete(key)
}

// deleteProcessedTime deletes the processedTime for a given height
func deleteProcessedTime(clientStore sdk.KVStore, height exported.Height) {
	key := ibctm.ProcessedTimeKey(height)
	clientStore.Delete(key)
}

// deleteProcessedHeight deletes the processedHeight for a given height
func deleteProcessedHeight(clientStore sdk.KVStore, height exported.Height) {
	key := ibctm.ProcessedHeightKey(height)
	clientStore.Delete(key)
}

// deleteIterationKey deletes the iteration key for a given height
func deleteIterationKey(clientStore sdk.KVStore, height exported.Height) {
	key := ibctm.IterationKey(height)
	clientStore.Delete(key)
}
