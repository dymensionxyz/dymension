package types

import (
	"bytes"
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/types"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// CheckCompatibility checks if the IBC state and Rollapp state are compatible
func CheckCompatibility(ibcState IBCState, raState RollappState) error {
	// Check if block descriptor state root matches IBC block header app hash
	if !bytes.Equal(ibcState.Root, raState.BlockDescriptor.StateRoot) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor state root does not match tendermint header app hash")
	}
	// Check if the validator pubkey matches the sequencer pubkey
	valHashFromStateInfo, err := GetValHashForSequencer(raState.BlockSequencer)
	if err != nil {
		return errors.Join(ibcclienttypes.ErrInvalidConsensus, err)
	}
	// The ValidatorsHash is only available when the block header is submitted (i.e only during MsgUpdateClient)
	if len(ibcState.ValidatorsHash) > 0 && !bytes.Equal(ibcState.ValidatorsHash, valHashFromStateInfo) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "validator does not match the sequencer")
	}
	if len(raState.NextBlockSequencer) == 0 {
		return ErrNextBlockDescriptorMissing
	}
	// Check if the nextValidatorHash matches for the sequencer for h+1 block descriptor
	nextValHashFromStateInfo, err := GetValHashForSequencer(raState.NextBlockSequencer)
	if err != nil {
		return errors.Join(ibcclienttypes.ErrInvalidConsensus, err)
	}
	if !bytes.Equal(ibcState.NextValidatorsHash, nextValHashFromStateInfo) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "next validator hash does not match the sequencer for h+1")
	}
	// Check if block descriptor timestamp is not present - this happens if the rollapp has not upgraded yet
	if raState.BlockDescriptor.Timestamp.IsZero() {
		return ErrTimestampNotFound
	}
	// Check if block descriptor timestamp matches IBC header timestamp
	if !ibcState.Timestamp.Equal(raState.BlockDescriptor.Timestamp) {
		return errorsmod.Wrap(ibcclienttypes.ErrInvalidConsensus, "block descriptor timestamp does not match tendermint header timestamp")
	}
	return nil
}

// GetValHashForSequencer creates a dummy tendermint validatorset to
// calculate the nextValHash for the sequencer and returns it
func GetValHashForSequencer(sequencerTmPubKeyBz []byte) ([]byte, error) {
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
	// Root is the app root shared by the IBC consensus state
	Root []byte
	// Height is the block height of the IBC consensus state the root is at
	Height uint64
	// ValidatorsHash is the tendermint pubkey of signer of the block header
	ValidatorsHash []byte
	// NextValidatorsHash is the hash of the next validator set for the next block
	NextValidatorsHash []byte
	// Timestamp is the block timestamp of the header
	Timestamp time.Time
}

type RollappState struct {
	// BlockSequencer is the tendermint pubkey of the sequencer who submitted the block descriptor for the required height
	BlockSequencer []byte
	// BlockDescriptor is the block descriptor for the required height
	BlockDescriptor rollapptypes.BlockDescriptor
	// NextBlockSequencer is the tendermint pubkey of the sequencer who submitted the block descriptor for the next height (h+1)
	NextBlockSequencer []byte
	// NextBlockDescriptor is the block descriptor for the next height (h+1)
	NextBlockDescriptor rollapptypes.BlockDescriptor
}
