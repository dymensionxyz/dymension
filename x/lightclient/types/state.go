package types

import (
	"bytes"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	errorsmod "cosmossdk.io/errors"
)

func StateCompatible(consensusState ibctm.ConsensusState, rollappBD rollapptypes.BlockDescriptor) error {
	// Check if block descriptor state root matches tendermint consensus state root
	// if bytes.Equal(rollappBD.StateRoot, consensusState.GetRoot().GetHash()) {
	// 	return nil
	// }
	// Check if block descriptor timestamp matches tendermint consensus state timestamp
	// if blockDescriptor.Timestamp != tendermintConsensusState.GetTimestamp() {
	// 	return
	// }

	// Check if the validator set hash matches the sequencer
	// if len(tendermintConsensusState.GetNextValidators().Validators) == 1 && tendermintConsensusState.GetNextValidators().Validators[0].Address {
	// 	return
	// }

	// Check if the nextValidatorHash matches the sequencer for h+1
	// todo: pass in all the stateinfo so we can lookup h+1 block descriptor
	return nil
}

func HeaderCompatible(header ibctm.Header, stateInfo rollapptypes.StateInfo) error {
	height := header.GetHeight()
	currentHeaderBD := stateInfo.GetBDs().BD[height.GetRevisionHeight()-stateInfo.GetStartHeight()]
	// Check if block descriptor state root matches tendermint header app hash
	if !bytes.Equal(currentHeaderBD.StateRoot, header.Header.AppHash) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor state root does not match tendermint header app hash")
	}
	if currentHeaderBD.Timestamp.Equal(header.Header.Time) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor timestamp does not match tendermint header timestamp")
	}
	// Check if the validator set hash matches the sequencer
	if len(header.ValidatorSet.Validators) == 1 && (string(header.ValidatorSet.Validators[0].Address) == stateInfo.Sequencer) { // todo: do proper data transformation before checking
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "validator set hash does not match the sequencer")
	}
	// Check if the nextValidatorHash matches the sequencer for h+1
	nextBlockDescriptor := stateInfo.GetBDs().BD[height.GetRevisionHeight()-stateInfo.GetStartHeight()+1]
	//if nextBlockDescriptor is part of same state info, the same sequencer
	if !bytes.Equal([]byte(stateInfo.Sequencer), header.Header.NextValidatorsHash) { // todo: do proper data transformation before checking
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "next validator hash does not match the sequencer for h+1")
	}

	return nil
}
