package keeper

import (
	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v8/modules/core/23-commitment/types"

	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
)

func (hook rollappHook) OnHardFork(ctx sdk.Context, rollappId string, lastValidHeight uint64) error {
	return hook.k.RollbackCanonicalClient(ctx, rollappId, lastValidHeight)
}

func (k Keeper) RollbackCanonicalClient(ctx sdk.Context, rollappId string, lastValidHeight uint64) error {
	client, found := k.GetCanonicalClient(ctx, rollappId)
	if !found {
		return nil
	}
	cs := k.ibcClientKeeper.ClientStore(ctx, client)

	var lastConsStateHeight exported.Height
	// iterate over all consensus states and metadata in the client store
	IterateConsensusStateDescending(cs, func(h exported.Height) bool {
		// iterate until we pass the new revision height
		if h.GetRevisionHeight() <= lastValidHeight {
			lastConsStateHeight = h
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

	// will be unfrozen on next state update
	if err := k.freezeClient(cs, lastConsStateHeight); err != nil {
		return errorsmod.Wrap(err, "freeze client")
	}

	return nil
}

// ResolveHardFork resolves the hard fork by resetting the client to the valid state
// and adding consensus states based on the block descriptors
// CONTRACT: canonical client is already set, state info exists
func (k Keeper) ResolveHardFork(ctx sdk.Context, rollappID string) error {
	clientID, _ := k.GetCanonicalClient(ctx, rollappID)                // already checked in the caller
	stateinfo, _ := k.rollappKeeper.GetLatestStateInfo(ctx, rollappID) // already checked in the caller

	err := k.UpdateClientFromStateInfo(ctx, clientID, &stateinfo)
	if err != nil {
		return errorsmod.Wrap(err, "update client from state info")
	}
	return nil
}

// UpdateClientFromStateInfo sets the consensus state from the state info
// and sets the metadata for the consensus state
// uses the latest height of the state info to set the client state
// CONTRACT: canonical client is already set, state info exists
func (k Keeper) UpdateClientFromStateInfo(ctx sdk.Context, clientID string, stateInfo *rollapptypes.StateInfo) error {
	clientStore := k.ibcClientKeeper.ClientStore(ctx, clientID)

	bd, _ := stateInfo.LastBlockDescriptor()
	height := bd.Height

	// sanity check
	clientLatestHeight := getClientStateTM(clientStore, k.cdc).GetLatestHeight().GetRevisionHeight()
	if height <= clientLatestHeight {
		return gerrc.ErrInternal.Wrapf("client latest height not less than new latest height: new: %d, client: %d",
			height, clientLatestHeight,
		)
	}

	proposer, _ := k.SeqK.RealSequencer(ctx, stateInfo.NextProposer)
	valHash, _ := proposer.ValsetHash()

	// add consensus states based on the block descriptors
	cs := ibctm.ConsensusState{
		Timestamp:          bd.Timestamp,
		Root:               commitmenttypes.NewMerkleRoot(bd.StateRoot),
		NextValidatorsHash: valHash,
	}

	setConsensusState(clientStore, k.cdc, clienttypes.NewHeight(1, height), &cs)
	setConsensusMetadata(ctx, clientStore, clienttypes.NewHeight(1, height))

	k.updateClientState(ctx, clientStore, height)
	return nil
}

// freezeClient freezes the client by setting the frozen height to the current height
func (k Keeper) freezeClient(clientStore storetypes.KVStore, heightI exported.Height) error {
	tmClientState := getClientStateTM(clientStore, k.cdc)

	// It's not fundamentally important to have a consensus state for the latest height (since
	// it can happen in normal operation due to IBC pruning) but we do best effort, because
	// ibctesting doesn't like not having it.
	height, ok := heightI.(clienttypes.Height)
	if !ok {
		return gerrc.ErrInternal.Wrap("height nil or not tm client height")
	}
	tmClientState.LatestHeight = height

	tmClientState.FrozenHeight = ibctm.FrozenHeight

	setClientState(clientStore, k.cdc, tmClientState)

	return nil
}

// updateClientState updates the client state by setting the latest height
func (k Keeper) updateClientState(ctx sdk.Context, clientStore storetypes.KVStore, height uint64) {
	tmClientState := getClientStateTM(clientStore, k.cdc)

	// set the latest height
	tmClientState.LatestHeight = clienttypes.NewHeight(1, height)
	tmClientState.FrozenHeight = clienttypes.ZeroHeight() // just to be sure

	setClientState(clientStore, k.cdc, tmClientState)

	// prune the oldest consensus state (similar to the ibc-go vanilla UpdateState flow)
	pruneOldestConsensusState(ctx, k.cdc, clientStore, *tmClientState)
}
