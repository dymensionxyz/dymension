package types

import (
	"bytes"
	"time"

	errorsmod "cosmossdk.io/errors"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CheckCompatibility(ibcState IBCState, raState RollappState) error {
	// Check if block descriptor state root matches IBC header app hash
	if !bytes.Equal(ibcState.Root, raState.BlockDescriptor.StateRoot) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor state root does not match tendermint header app hash")
	}
	// Check if block descriptor timestamp matches IBC header timestamp
	if !ibcState.Timestamp.Equal(raState.BlockDescriptor.Timestamp) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor timestamp does not match tendermint header timestamp")
	}
	// in case of msgcreateclinet validator info is not available
	// Check if the validator set hash matches the sequencer
	if len(ibcState.Validator) == 1 && bytes.Equal(ibcState.Validator, []byte(raState.BlockSequencer)) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "validator set hash does not match the sequencer")
	}

	// Check if the nextValidatorHash matches the sequencer for h+1
	if !bytes.Equal(ibcState.NextValidatorsHash, []byte(raState.NextBlockSequencer)) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "next validator hash does not match the sequencer for h+1")
	}
	return nil
}

type IBCState struct {
	Root               []byte
	Height             uint64
	Validator          []byte
	NextValidatorsHash []byte
	Timestamp          time.Time
}

type RollappState struct {
	BlockSequencer      string
	BlockDescriptor     rollapptypes.BlockDescriptor
	NextBlockSequencer  string
	NextBlockDescriptor rollapptypes.BlockDescriptor
}
