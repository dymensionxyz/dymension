package keeper

import (
	"encoding/binary"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (hook rollappHook) OnHardFork(ctx sdk.Context, rollappId string, height uint64) error {
	return hook.k.RollbackClient(ctx, rollappId, height)
}

func (k Keeper) RollbackClient(ctx sdk.Context, client string, height uint64) error {
	cs := k.ibcClientKeeper.ClientStore(ctx, client)

	// iterate over all consensus states and metadata in the client store
	ibctm.IterateConsensusStateAscending(cs, func(h exported.Height) bool {
		// if the height is lower than the target height, continue
		if h.GetRevisionHeight() < height {
			return false
		}

		// delete consensus state and metadata
		deleteConsensusState(cs, h)
		deleteConsensusMetadata(cs, h)

		// clean the optimistic updates valset
		k.RemoveConsensusStateValHash(ctx, client, height)

		// FIXME: marks the hardfork height
		// k.markHardForkHeight(ctx, targetHeight)
		return false
	})

	// reset IBC client
	err := k.resetClientToLastValidState(ctx, client)
	if err != nil {
		return errorsmod.Wrap(err, "failed to reset client to last valid state")
	}

	return nil
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

func (k Keeper) resetClientToLastValidState(ctx sdk.Context, client string) error {
	c, ok := k.ibcClientKeeper.GetClientState(ctx, client)
	if !ok {
		return types.ErrorMissingClientState
	}

	// Cast client state to tendermint client state
	tmClientState, ok := c.(*ibctm.ClientState)
	if !ok {
		return types.ErrorInvalidClientType
	}

	// update the height of the client to the last valid height
	prevHeight, found := GetPreviousConsensusStateHeight(k.ibcClientKeeper.ClientStore(ctx, client), k.cdc, tmClientState.LatestHeight)
	if !found {
		// if no previous height is found, set the height to 0
		prevHeight = clienttypes.Height{RevisionNumber: 0, RevisionHeight: 0}
	}

	// set the client state to the previous state
	tmClientState.LatestHeight = prevHeight.(clienttypes.Height)
	setClientState(k.ibcClientKeeper.ClientStore(ctx, client), k.cdc, tmClientState)

	return nil
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

// setClientState stores the client state
func setClientState(clientStore sdk.KVStore, cdc codec.BinaryCodec, clientState exported.ClientState) {
	key := host.ClientStateKey()
	val := clienttypes.MustMarshalClientState(cdc, clientState)
	clientStore.Set(key, val)
}

func bigEndianHeightBytes(height exported.Height) []byte {
	heightBytes := make([]byte, 16)
	binary.BigEndian.PutUint64(heightBytes, height.GetRevisionNumber())
	binary.BigEndian.PutUint64(heightBytes[8:], height.GetRevisionHeight())
	return heightBytes
}
