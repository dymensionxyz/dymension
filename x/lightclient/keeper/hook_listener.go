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
	if !hook.k.Enabled() {
		return nil
	}

	client, ok := hook.k.GetCanonicalClient(ctx, rollappId)
	if !ok {
		client, ok = hook.k.GetProspectiveCanonicalClient(ctx, rollappId, stateInfo.GetLatestHeight()-1)
		if ok {
			hook.k.SetCanonicalClient(ctx, rollappId, client)
		}
		return nil
	}

	// first state after hardfork, should reset the client to active state
	if hook.k.IsHardForkingInProgress(ctx, rollappId) {
		err := hook.k.ResolveHardFork(ctx, rollappId)
		if err != nil {
			return errorsmod.Wrap(err, "resolve hard fork")
		}
		return nil
	}

	// TODO: check hard fork in progress here

	seq, err := hook.k.SeqK.RealSequencer(ctx, stateInfo.Sequencer)
	if err != nil {
		return errorsmod.Wrap(errors.Join(gerrc.ErrInternal, err), "get sequencer for state info")
	}

	// [hStart-1..,hEnd) is correct because we compare against a next validators hash
	for h := stateInfo.GetStartHeight() - 1; h < stateInfo.GetLatestHeight(); h++ {
		if err := hook.validateOptimisticUpdate(ctx, rollappId, client, seq, stateInfo, h); err != nil {
			return errorsmod.Wrap(err, "validate optimistic update")
		}
	}

	// we now verified everything up to and including stateInfo.GetLatestHeight()-1
	// so we should prune everything up to stateInfo.GetLatestHeight()-1
	// this removes the unbonding condition for the sequencers
	if err := hook.k.PruneSignersBelow(ctx, client, stateInfo.GetLatestHeight()); err != nil {
		return errorsmod.Wrap(err, "prune signers")
	}

	return nil
}

func (hook rollappHook) validateOptimisticUpdate(
	ctx sdk.Context,
	rollapp string,
	client string,
	nextSequencer sequencertypes.Sequencer,
	cache *rollapptypes.StateInfo, // a place to look up the BD for a height
	h uint64,
) error {
	got, ok := hook.getConsensusState(ctx, client, h)
	if !ok {
		// done, nothing to validate
		return nil
	}
	expectBD, err := hook.getBlockDescriptor(ctx, rollapp, cache, h)
	if err != nil {
		return err
	}
	expect := types.RollappState{
		BlockDescriptor:    expectBD,
		NextBlockSequencer: nextSequencer,
	}

	err = types.CheckCompatibility(*got, expect)
	if err != nil {
		return errors.Join(gerrc.ErrFault, err)
	}

	// everything is fine
	return nil
}

func (hook rollappHook) getBlockDescriptor(ctx sdk.Context,
	rollapp string,
	cache *rollapptypes.StateInfo,
	h uint64,
) (rollapptypes.BlockDescriptor, error) {
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
	h uint64,
) (*ibctm.ConsensusState, bool) {
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
