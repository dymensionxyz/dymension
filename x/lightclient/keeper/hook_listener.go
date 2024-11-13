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

	client, ok := hook.k.GetCanonicalClient(ctx, rollappId)
	if !ok {
		client, ok = hook.k.FindMatchingClient(ctx, stateInfo)
		if ok {
			hook.k.SetCanonicalClient(ctx, rollappId, client)
			if err := hook.k.PruneSignersBelow(ctx, client, stateInfo.GetLatestHeight()+1); err != nil {
				return errorsmod.Wrap(err, "prune signers")
			}
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

	if err := hook.validateOptimisticUpdate(ctx, client, stateInfo); err != nil {
		return errorsmod.Wrap(err, "validate optimistic update")
	}

	// we now verified everything up to and including stateInfo.GetLatestHeight()
	// this removes the unbonding condition for the sequencers
	if err := hook.k.PruneSignersBelow(ctx, client, stateInfo.GetLatestHeight()+1); err != nil {
		return errorsmod.Wrap(err, "prune signers")
	}

	return nil
}

func (hook rollappHook) validateOptimisticUpdate(
	ctx sdk.Context,
	client string,
	stateInfo *rollapptypes.StateInfo, // a place to look up the BD for a height
) error {
	for h := stateInfo.GetStartHeight(); h <= stateInfo.GetLatestHeight(); h++ {
		got, ok := hook.getConsensusState(ctx, client, h)
		if !ok {
			continue
		}

		err := hook.k.ValidateUpdatePessimistically(ctx, stateInfo, got, h)
		if err != nil {
			return errorsmod.Wrapf(err, "validate pessimistic h: %d", h)
		}
	}

	// everything is fine
	return nil
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
