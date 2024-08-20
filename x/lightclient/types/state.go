package types

import (
	"bytes"
	"time"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/types"

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
	// in case of msgcreateclient, validator info is not available. it is only send in msgupdateclient as header info
	// Check if the validator set hash matches the sequencer
	if len(ibcState.Validator) > 0 && !bytes.Equal(ibcState.Validator, raState.BlockSequencer) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "validator set hash does not match the sequencer")
	}

	// Check if the nextValidatorHash matches the sequencer for h+1
	nextValHashFromStateInfo, err := getValHashForSequencer(raState.NextBlockSequencer)
	if err != nil {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, err.Error())
	}
	if !bytes.Equal(ibcState.NextValidatorsHash, nextValHashFromStateInfo) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "next validator hash does not match the sequencer for h+1")
	}
	return nil
}

func getValHashForSequencer(sequencerTmPubKeyBz []byte) ([]byte, error) {
	var tmpk tmprotocrypto.PublicKey
	err := tmpk.Unmarshal(sequencerTmPubKeyBz)
	if err != nil {
		return nil, err
	}
	var nextValSet cmttypes.ValidatorSet
	updates, err := cmttypes.PB2TM.ValidatorUpdates([]abci.ValidatorUpdate{{Power: 1, PubKey: tmpk}})
	if err != nil {
		return nil, err
	}
	err = nextValSet.UpdateWithChangeSet(updates)
	if err != nil {
		return nil, err
	}
	return nextValSet.Hash(), nil
}

type IBCState struct {
	Root               []byte
	Height             uint64
	Validator          []byte
	NextValidatorsHash []byte
	Timestamp          time.Time
}

type RollappState struct {
	BlockSequencer      []byte
	BlockDescriptor     rollapptypes.BlockDescriptor
	NextBlockSequencer  []byte
	NextBlockDescriptor rollapptypes.BlockDescriptor
}
