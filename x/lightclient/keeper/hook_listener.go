package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ rollapptypes.RollappHooks = rollappHook{}

// Hooks wrapper struct for rollapp keeper.
type rollappHook struct {
	rollapptypes.StubRollappCreatedHooks
	k Keeper
}

// RollappHooks returns the wrapper struct.
func (k Keeper) RollappHooks() rollapptypes.RollappHooks {
	return rollappHook{k: k}
}

// AfterUpdateState is called after a state update is made to a rollapp.
// This hook checks if the rollapp has a canonical IBC light client and if the Consensus state is compatible with the state update
// and punishes the sequencer if it is not
func (hook rollappHook) AfterUpdateState(
	ctx sdk.Context,
	rollappId string,
	stateInfo *rollapptypes.StateInfo,
) error {
	client, ok := hook.k.GetCanonicalClient(ctx, rollappId)
	if !ok {
		client, ok = hook.k.GetProspectiveCanonicalClient(ctx, rollappId, stateInfo.GetLatestHeight()-1)
		if ok {
			hook.k.SetCanonicalClient(ctx, rollappId, client)
		}
		return nil
	}

	seq, err := hook.k.rollappKeeper.GetRealSequencer(ctx, stateInfo.Sequencer)
	if err != nil {
		return gerrc.ErrInternal.Wrap("get sequencer for state info")
	}

	// [h-1..,h) is correct because we compare against a next validators hash
	for h := stateInfo.GetLatestHeight() - 1; h < stateInfo.GetLatestHeight(); h++ {
		if err := hook.validateOptimisticUpdate(ctx, rollappId, client, seq, stateInfo, h); err != nil {
			if errorsmod.IsOf(err, gerrc.ErrFault) {
				// TODO: should double check this flow when implementing hard fork
				break
			}
			return err
		}
	}
	return nil
}

func (hook rollappHook) validateOptimisticUpdate(
	ctx sdk.Context,
	rollapp string,
	client string,
	nextSequencer sequencertypes.Sequencer,
	cache *rollapptypes.StateInfo,
	h uint64,
) error {
	expectBD, err := hook.getBlockDescriptor(ctx, rollapp, cache, h)
	if err != nil {
		// TODO: not found?
		return err
	}
	expect := types.RollappState{
		BlockDescriptor:    expectBD,
		NextBlockSequencer: nextSequencer,
	}
	got, ok := hook.getConsensusState(ctx, client, h)
	if !ok {
		// done
		return nil
	}
	signerAddr, err := hook.k.GetSigner(ctx, client, h)
	if err != nil {
		return gerrc.ErrInternal.Wrap("got cons state but no signer addr")
	}
	signer, err := hook.k.sequencerKeeper.GetRealSequencer(ctx, signerAddr)
	if err != nil {
		return gerrc.ErrInternal.Wrap("got cons state but no signer seq")
	}
	err = hook.k.RemoveSigner(ctx, signer, client, h)
	if err != nil {
		return errorsmod.Wrap(err, "remove signer")
	}
	err = types.CheckCompatibility(*got, expect)
	if err == nil {
		// everything is fine
		return nil
	}
	// fraud!
	err = hook.k.rollappKeeper.HandleFraud(ctx, signer.RollappId, client, h, signer.Address)
	if err != nil {
		return errorsmod.Wrap(err, "handle fraud")
	}
	return gerrc.ErrFault
}

func (hook rollappHook) getBlockDescriptor(ctx sdk.Context,
	rollapp string,
	cache *rollapptypes.StateInfo,
	h uint64) (rollapptypes.BlockDescriptor, error) {
	stateInfo := cache
	if !stateInfo.ContainsHeight(h) {
		var err error
		stateInfo, err = hook.k.rollappKeeper.FindStateInfoByHeight(ctx, rollapp, h)
		if err != nil {
			return rollapptypes.BlockDescriptor{}, errors.Join(err, gerrc.ErrInternal)
		}
	}
	bd, _ := stateInfo.GetBlockDescriptor(h)
	return bd, nil
}

func (hook rollappHook) getConsensusState(ctx sdk.Context,
	client string,
	h uint64) (*ibctm.ConsensusState, bool) {
	cs, _ := hook.k.ibcClientKeeper.GetClientState(ctx, client)
	height := ibcclienttypes.NewHeight(cs.GetLatestHeight().GetRevisionNumber(), h)
	consensusState, ok := hook.k.ibcClientKeeper.GetClientConsensusState(ctx, client, height)
	if !ok {
		return nil, false
	}
	tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
	if !ok {
		return nil, false
	}
	return tmConsensusState, true

}

func (hook rollappHook) checkStateForHeight(ctx sdk.Context,
	rollappId string,
	bd rollapptypes.BlockDescriptor,
	canonicalClient string,
	seq sequencertypes.Sequencer,
	blockValHash []byte,
) error {
	cs, _ := hook.k.ibcClientKeeper.GetClientState(ctx, canonicalClient)
	height := ibcclienttypes.NewHeight(cs.GetLatestHeight().GetRevisionNumber(), bd.GetHeight())
	consensusState, _ := hook.k.ibcClientKeeper.GetClientConsensusState(ctx, canonicalClient, height)
	// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
	tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
	if !ok {
		return nil
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: seq,
	}
	sequencerAddress, err := hook.k.GetSigner(ctx, canonicalClient, bd.GetHeight())
	if err != nil {
		return err
	}
	err = types.CheckCompatibility(*tmConsensusState, rollappState)
	if err != nil {
		// If the state is not compatible,
		// Take this state update as source of truth over the IBC update
		// Punish the block proposer of the IBC signed header

		err = hook.k.rollappKeeper.HandleFraud(ctx, rollappId, canonicalClient, bd.GetHeight(), sequencerAddress)
		if err != nil {
			return err
		}
	}
	hook.k.RemoveSigner(ctx)
	hook.k.RemoveConsensusStateValHash(ctx, canonicalClient, bd.GetHeight())
	return nil
}

func legacy() {
	latestHeight := stateInfo.GetLatestHeight()
	// We check from latestHeight-1 downwards, as the nextValHash for latestHeight will not be available until next stateupdate
	for h := latestHeight - 1; h >= stateInfo.StartHeight; h-- {
		bd, _ := stateInfo.GetBlockDescriptor(h)
		// Check if any optimistic updates were made for the given height
		blockValHash, found := hook.k.GetConsensusStateValHash(ctx, client, bd.GetHeight())
		if !found {
			continue
		}
		err := hook.checkStateForHeight(ctx, rollappId, bd, client, seq, blockValHash)
		if err != nil {
			return err
		}
	}
	// Check for the last BD from the previous stateInfo as now we have the nextValhash available for that block
	blockValHash, ok := hook.k.GetConsensusStateValHash(ctx, client, stateInfo.StartHeight-1)
	if ok {
		previousStateInfo, err := hook.k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, stateInfo.StartHeight-1)
		if err != nil {
			return err
		}
		bd, _ := previousStateInfo.GetBlockDescriptor(stateInfo.StartHeight - 1)
		err = hook.checkStateForHeight(ctx, rollappId, bd, client, seq, blockValHash)
		if err != nil {
			return err
		}
	}
	return nil
}
