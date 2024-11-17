package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

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

	client, canonical := hook.k.GetCanonicalClient(ctx, rollappId)
	if canonical {
		// validate state info against optimistically accepted headers
		_, err := hook.k.ValidateOptimisticUpdates(ctx, client, stateInfo)
		if err != nil {
			return errorsmod.Wrap(err, "validate optimistic update")
		}
	} else {
		var ok bool
		client, ok = hook.k.FindPotentialClient(ctx, stateInfo)
		if !ok {
			return nil
		}
		hook.k.SetCanonicalClient(ctx, rollappId, client)
	}

	// we now verified everything up to and including stateInfo.GetLatestHeight()
	// this removes the unbonding condition for the sequencers
	if err := hook.k.PruneSignersBelow(ctx, client, stateInfo.GetLatestHeight()+1); err != nil {
		return errorsmod.Wrap(err, "prune signers")
	}

	// first state after hardfork, should reset the client to active state
	if hook.k.IsHardForkingInProgress(ctx, rollappId) {
		err := hook.k.ResolveHardFork(ctx, rollappId)
		if err != nil {
			return errorsmod.Wrap(err, "resolve hard fork")
		}
		return nil
	}

	return nil
}

func (k Keeper) ValidateOptimisticUpdates(
	ctx sdk.Context,
	client string,
	stateInfo *rollapptypes.StateInfo, // a place to look up the BD for a height
) (matched bool, err error) {
	atLeastOneMatch := false
	for h := stateInfo.GetStartHeight(); h <= stateInfo.GetLatestHeight(); h++ {
		got, ok := k.getConsensusState(ctx, client, h)
		if !ok {
			continue
		}

		err := k.ValidateHeaderAgainstStateInfo(ctx, stateInfo, got, h)
		if err != nil {
			return false, errorsmod.Wrapf(err, "validate pessimistic h: %d", h)
		}

		atLeastOneMatch = true
	}

	// everything is fine
	return atLeastOneMatch, nil
}

func (k Keeper) getConsensusState(ctx sdk.Context,
	client string,
	h uint64,
) (*ibctm.ConsensusState, bool) {
	cs, _ := k.ibcClientKeeper.GetClientState(ctx, client)
	height := ibcclienttypes.NewHeight(cs.GetLatestHeight().GetRevisionNumber(), h)
	consensusState, ok := k.ibcClientKeeper.GetClientConsensusState(ctx, client, height)
	if !ok {
		return nil, false
	}
	tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
	if !ok {
		return nil, false
	}
	return tmConsensusState, true
}
