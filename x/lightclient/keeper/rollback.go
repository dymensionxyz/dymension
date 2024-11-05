package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"

	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (hook rollappHook) OnHardFork(ctx sdk.Context, rollappId string, fraudHeight uint64) error {
	hook.k.RollbackCanonicalClient(ctx, rollappId, fraudHeight)
	return nil
}

func (k Keeper) RollbackCanonicalClient(ctx sdk.Context, rollappId string, fraudHeight uint64) {
	client, found := k.GetCanonicalClient(ctx, rollappId)
	if !found {
		k.Logger(ctx).Error("Canonical client not found", "rollappId", rollappId)
		return
	}
	cs := k.ibcClientKeeper.ClientStore(ctx, client)

	// iterate over all consensus states and metadata in the client store
	IterateConsensusStateDescending(cs, func(h exported.Height) bool {
		// iterate until we pass the fraud height
		if h.GetRevisionHeight() < fraudHeight {
			return true
		}

		// delete consensus state and metadata
		deleteConsensusState(cs, h)
		deleteConsensusMetadata(cs, h)

		return false
	})

	// clean the optimistic updates valset
	k.PruneSigners(ctx, client, fraudHeight-1)

	// marks that hard fork is in progress
	k.setHardForkInProgress(ctx, rollappId)

	// freeze the client
	// it will be released after the hardfork is resolved (on the next state update)
	k.freezeClient(cs, fraudHeight)
}

// set latest IBC consensus state nextValHash to the current proposing sequencer.
func (k Keeper) ResolveHardFork(ctx sdk.Context, rollappID string) {
	client, _ := k.GetCanonicalClient(ctx, rollappID)
	clientStore := k.ibcClientKeeper.ClientStore(ctx, client)

	stateinfo, _ := k.rollappKeeper.GetLatestStateInfo(ctx, rollappID)
	height := stateinfo.GetLatestHeight()
	bd := stateinfo.GetLatestBlockDescriptor()

	// get the valHash of this sequencer
	proposer, _ := k.SeqK.RealSequencer(ctx, stateinfo.Sequencer)
	valHash, _ := proposer.ValsetHash()

	// unfreeze the client and set the latest height
	k.resetClientToValidState(clientStore, height)
	// add consensus states based on the block descriptors
	cs := ibctm.ConsensusState{
		Timestamp:          bd.Timestamp,
		Root:               commitmenttypes.NewMerkleRoot(bd.StateRoot),
		NextValidatorsHash: valHash,
	}

	setConsensusState(clientStore, k.cdc, clienttypes.NewHeight(1, height), &cs)

	k.setHardForkResolved(ctx, rollappID)
}

// freezeClient freezes the client by setting the frozen height to the current height
func (k Keeper) freezeClient(clientStore sdk.KVStore, height uint64) {
	c := getClientState(clientStore, k.cdc)
	tmClientState, _ := c.(*ibctm.ClientState)

	// freeze the client
	tmClientState.FrozenHeight = clienttypes.NewHeight(1, height)
	setClientState(clientStore, k.cdc, tmClientState)
}

// freezeClient freezes the client by setting the frozen height to the current height
func (k Keeper) resetClientToValidState(clientStore sdk.KVStore, height uint64) {
	c := getClientState(clientStore, k.cdc)
	tmClientState, _ := c.(*ibctm.ClientState)

	// unfreeze the client and set the latest height
	tmClientState.FrozenHeight = clienttypes.ZeroHeight()
	tmClientState.LatestHeight = clienttypes.NewHeight(1, height)

	setClientState(clientStore, k.cdc, tmClientState)
}

func GetPreviousConsensusStateHeight(clientStore sdk.KVStore, cdc codec.BinaryCodec, height exported.Height) (exported.Height, bool) {
	iterateStore := prefix.NewStore(clientStore, []byte(ibctm.KeyIterateConsensusStatePrefix))
	iterator := iterateStore.ReverseIterator(nil, bigEndianHeightBytes(height))
	defer iterator.Close() // nolint: errcheck

	if !iterator.Valid() {
		return nil, false
	}

	prevHeight := ibctm.GetHeightFromIterationKey(iterator.Key())
	return prevHeight, true
}

// IterateConsensusStateDescending iterates through all consensus states in descending order
// until cb returns true.
func IterateConsensusStateDescending(clientStore sdk.KVStore, cb func(height exported.Height) (stop bool)) {
	iterator := sdk.KVStoreReversePrefixIterator(clientStore, []byte(ibctm.KeyIterateConsensusStatePrefix))
	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		iterKey := iterator.Key()
		height := ibctm.GetHeightFromIterationKey(iterKey)
		if cb(height) {
			break
		}
	}
}
