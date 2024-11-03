package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"

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
	// FIXME: iterate descending
	ibctm.IterateConsensusStateAscending(cs, func(h exported.Height) bool {
		// if the height is lower than the target height, continue
		if h.GetRevisionHeight() < fraudHeight {
			return false
		}

		// delete consensus state and metadata
		deleteConsensusState(cs, h)
		deleteConsensusMetadata(cs, h)

		// clean the optimistic updates valset
		k.RemoveConsensusStateValHash(ctx, client, h.GetRevisionHeight())

		return false
	})
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
	proposer, _ := k.sequencerKeeper.GetSequencer(ctx, stateinfo.Sequencer)
	valHash, _ := proposer.GetDymintPubKeyHash()

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

func (k Keeper) setHardForkInProgress(ctx sdk.Context, rollappID string) {
	ctx.KVStore(k.storeKey).Set(types.HardForkKey(rollappID), []byte{0x01})
}

// remove the hardfork key from the store
func (k Keeper) setHardForkResolved(ctx sdk.Context, rollappID string) {
	ctx.KVStore(k.storeKey).Delete(types.HardForkKey(rollappID))
}

// checks if rollapp is hard forking
func (k Keeper) IsHardForkingInProgress(ctx sdk.Context, rollappID string) bool {
	return ctx.KVStore(k.storeKey).Has(types.HardForkKey(rollappID))
}
