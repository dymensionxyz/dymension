package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"

	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

func (hook rollappHook) OnHardFork(ctx sdk.Context, rollappId string, lastValidHeight uint64) error {
	return hook.k.RollbackCanonicalClient(ctx, rollappId, lastValidHeight)
}

func (k Keeper) RollbackCanonicalClient(ctx sdk.Context, rollappId string, lastValidHeight uint64) error {
	client, found := k.GetCanonicalClient(ctx, rollappId)
	if !found {
		return gerrc.ErrFailedPrecondition.Wrap("canonical client not found")
	}
	cs := k.ibcClientKeeper.ClientStore(ctx, client)

	// iterate over all consensus states and metadata in the client store
	IterateConsensusStateDescending(cs, func(h exported.Height) bool {
		// iterate until we pass the new revision height
		if h.GetRevisionHeight() <= lastValidHeight {
			return true
		}

		// delete consensus state and metadata
		deleteConsensusState(cs, h)
		deleteConsensusMetadata(cs, h)

		return false
	})

	// we DO want to prune the signer of the last valid height:
	// the only reason we didn't do it before was because we were waiting for next validators hash
	// but now we don't care about that
	err := k.PruneSignersAbove(ctx, client, lastValidHeight-1)
	if err != nil {
		return errorsmod.Wrap(err, "prune signers above")
	}

	// freeze the client
	// it will be released after the hardfork is resolved (on the next state update)
	k.freezeClient(cs, lastValidHeight)

	return nil
}

// ResolveHardFork resolves the hard fork by resetting the client to the valid state
// and adding consensus states based on the block descriptors
// CONTRACT: canonical client is already set, state info exists
func (k Keeper) ResolveHardFork(ctx sdk.Context, rollappID string) error {
	clientID, _ := k.GetCanonicalClient(ctx, rollappID) // already checked in the caller
	clientStore := k.ibcClientKeeper.ClientStore(ctx, clientID)

	stateinfo, _ := k.rollappKeeper.GetLatestStateInfo(ctx, rollappID) // already checked in the caller

	height := stateinfo.StartHeight
	// sanity check
	client := getClientStateTM(clientStore, k.cdc)
	clientHeight := client.GetLatestHeight().GetRevisionHeight()
	if height <= clientHeight {
		return gerrc.ErrInternal.Wrapf("client latest height not less than new latest height: new: %d, client: %d",
			height, clientHeight,
		)
	}

	bd := stateinfo.BDs.BD[0]

	// get the valHash of this sequencer
	// we assume the proposer of the first state update after the hard fork won't be rotated in the next block
	proposer, _ := k.SeqK.RealSequencer(ctx, stateinfo.Sequencer)
	valHash, _ := proposer.ValsetHash()

	// add consensus states based on the block descriptors
	cs := ibctm.ConsensusState{
		Timestamp:          bd.Timestamp,
		Root:               commitmenttypes.NewMerkleRoot(bd.StateRoot),
		NextValidatorsHash: valHash,
	}

	setConsensusState(clientStore, k.cdc, clienttypes.NewHeight(1, height), &cs)
	setConsensusMetadata(ctx, clientStore, clienttypes.NewHeight(1, height))

	k.unfreezeClient(clientStore, height)

	return nil
}

// freezeClient freezes the client by setting the frozen height to the current height
func (k Keeper) freezeClient(clientStore sdk.KVStore, height uint64) {
	tmClientState := getClientStateTM(clientStore, k.cdc)

	// freeze the client
	tmClientState.FrozenHeight = ibctm.FrozenHeight
	tmClientState.LatestHeight = clienttypes.NewHeight(1, height)

	setClientState(clientStore, k.cdc, tmClientState)
}

// freezeClient freezes the client by setting the frozen height to the current height
func (k Keeper) unfreezeClient(clientStore sdk.KVStore, height uint64) {
	tmClientState := getClientStateTM(clientStore, k.cdc)

	// unfreeze the client and set the latest height
	tmClientState.FrozenHeight = clienttypes.ZeroHeight()
	tmClientState.LatestHeight = clienttypes.NewHeight(1, height)

	setClientState(clientStore, k.cdc, tmClientState)
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
