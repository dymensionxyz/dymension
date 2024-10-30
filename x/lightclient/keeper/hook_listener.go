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

	seq, err := hook.k.SeqK.GetRealSequencer(ctx, stateInfo.Sequencer)
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
		return err
	}
	expect := types.RollappState{
		BlockDescriptor:    expectBD,
		NextBlockSequencer: nextSequencer,
	}
	got, ok := hook.getConsensusState(ctx, client, h)
	if !ok {
		// done, nothing to validate
		return nil
	}
	signerAddr, err := hook.k.GetSigner(ctx, client, h)
	if err != nil {
		return gerrc.ErrInternal.Wrap("got cons state but no signer addr")
	}
	signer, err := hook.k.SeqK.GetRealSequencer(ctx, signerAddr)
	if err != nil {
		return gerrc.ErrInternal.Wrap("got cons state but no signer seq")
	}
	// remove to allow unbond
	err = hook.k.RemoveSigner(ctx, signer.Address, client, h)
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
