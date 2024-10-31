package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

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

func GetPreviousConsensusStateHeight(clientStore sdk.KVStore, cdc codec.BinaryCodec, height exported.Height) (exported.Height, bool) {
	iterateStore := prefix.NewStore(clientStore, []byte(ibctm.KeyIterateConsensusStatePrefix))
	iterator := iterateStore.ReverseIterator(nil, bigEndianHeightBytes(height))
	defer iterator.Close()

	if !iterator.Valid() {
		return nil, false
	}

	prevHeight := ibctm.GetHeightFromIterationKey(iterator.Key())
	return prevHeight, true
}

func bigEndianHeightBytes(height exported.Height) []byte {
	heightBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(heightBytes, height.GetRevisionNumber())
	binary.BigEndian.PutUint64(heightBytes[8:], height.GetRevisionHeight())
	return heightBytes
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
