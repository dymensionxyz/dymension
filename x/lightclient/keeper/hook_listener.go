package keeper

import (
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"

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
	ctx.Logger().Info("Light client. AfterUpdateState.") // TODO: remove
	canonicalClient, found := hook.k.GetCanonicalClient(ctx, rollappId)
	if !found {
		canonicalClient, foundClient := hook.k.GetProspectiveCanonicalClient(ctx, rollappId, stateInfo.GetLatestHeight()-1)
		if foundClient {
			hook.k.SetCanonicalClient(ctx, rollappId, canonicalClient)
			ctx.Logger().Info("Light client. Set canonical.", "canonical", canonicalClient) // TODO: remove
		}
		return nil
	}
	sequencerPk, err := hook.k.GetSequencerPubKey(ctx, stateInfo.Sequencer)
	if err != nil {
		return err
	}
	latestHeight := stateInfo.GetLatestHeight()
	// We check from latestHeight-1 downwards, as the nextValHash for latestHeight will not be available until next stateupdate
	for h := latestHeight - 1; h >= stateInfo.StartHeight; h-- {
		bd, _ := stateInfo.GetBlockDescriptor(h)
		// Check if any optimistic updates were made for the given height
		blockValHash, found := hook.k.GetConsensusStateValHash(ctx, canonicalClient, bd.GetHeight())
		if !found {
			continue
		}
		err := hook.checkStateForHeight(ctx, rollappId, bd, canonicalClient, sequencerPk, blockValHash)
		if err != nil {
			return err
		}
	}
	// Check for the last BD from the previous stateInfo as now we have the nextValhash available for that block
	blockValHash, found := hook.k.GetConsensusStateValHash(ctx, canonicalClient, stateInfo.StartHeight-1)
	if found {
		previousStateInfo, err := hook.k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, stateInfo.StartHeight-1)
		if err != nil {
			return err
		}
		bd, _ := previousStateInfo.GetBlockDescriptor(stateInfo.StartHeight - 1)
		err = hook.checkStateForHeight(ctx, rollappId, bd, canonicalClient, sequencerPk, blockValHash)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hook rollappHook) checkStateForHeight(ctx sdk.Context, rollappId string, bd rollapptypes.BlockDescriptor, canonicalClient string, sequencerPk tmprotocrypto.PublicKey, blockValHash []byte) error {
	cs, _ := hook.k.ibcClientKeeper.GetClientState(ctx, canonicalClient)
	height := ibcclienttypes.NewHeight(cs.GetLatestHeight().GetRevisionNumber(), bd.GetHeight())
	consensusState, _ := hook.k.ibcClientKeeper.GetClientConsensusState(ctx, canonicalClient, height)
	// Cast consensus state to tendermint consensus state - we need this to check the state root and timestamp and nextValHash
	tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
	if !ok {
		return nil // TODO: why nil?
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: sequencerPk,
	}
	err := types.CheckCompatibility(*tmConsensusState, rollappState)
	if err != nil {
		// If the state is not compatible,
		// Take this state update as source of truth over the IBC update
		// Punish the block proposer of the IBC signed header
		sequencerAddress, err := hook.k.GetSequencerFromValHash(ctx, rollappId, blockValHash)
		if err != nil {
			return err
		}
		err = hook.k.rollappKeeper.HandleFraud(ctx, rollappId, canonicalClient, bd.GetHeight(), sequencerAddress)
		if err != nil {
			return err
		}
	}
	hook.k.RemoveConsensusStateValHash(ctx, canonicalClient, bd.GetHeight())
	return nil
}
