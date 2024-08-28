package types

import (
	"bytes"
	"errors"
	"time"

	errorsmod "cosmossdk.io/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cmttypes "github.com/cometbft/cometbft/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// CheckCompatibility checks if the IBC state and Rollapp state are compatible
// Compatibility Criteria:
// 1. The app root shared by the IBC consensus state matches the block descriptor state root for the same height
// 2. The next validator hash shared by the IBC consensus state matches the sequencer hash for the next block descriptor
// 3. The block descriptor timestamp matches the tendermint header timestamp
func CheckCompatibility(ibcState IBCState, raState RollappState) error {
	// Check if block descriptor state root matches IBC block header app hash
	if !bytes.Equal(ibcState.Root, raState.BlockDescriptor.StateRoot) {
		return errorsmod.Wrap(ErrStateRootsMismatch, "block descriptor state root does not match tendermint header app hash")
	}
	// Check if the nextValidatorHash matches for the sequencer for h+1 block descriptor
	nextValHashFromStateInfo, err := GetValHashForSequencer(raState.NextBlockSequencer)
	if err != nil {
		return errors.Join(ErrValidatorHashMismatch, err)
	}
	if !bytes.Equal(ibcState.NextValidatorsHash, nextValHashFromStateInfo) {
		return errorsmod.Wrap(ErrValidatorHashMismatch, "next validator hash does not match the sequencer for h+1")
	}
	if !raState.BlockDescriptor.Timestamp.IsZero() && !ibcState.Timestamp.Equal(raState.BlockDescriptor.Timestamp) {
		return errorsmod.Wrap(ErrTimestampMismatch, "block descriptor timestamp does not match tendermint header timestamp")
	}
	return nil
}

// GetValHashForSequencer creates a dummy tendermint validatorset to
// calculate the nextValHash for the sequencer and returns it
func GetValHashForSequencer(sequencerTmPubKey tmprotocrypto.PublicKey) ([]byte, error) {
	var nextValSet cmttypes.ValidatorSet
	updates, err := cmttypes.PB2TM.ValidatorUpdates([]abci.ValidatorUpdate{{Power: 1, PubKey: sequencerTmPubKey}})
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
	// NextValidatorsHash is the hash of the next validator set for the next block
	NextValidatorsHash []byte
	// Timestamp is the block timestamp of the header
	Timestamp time.Time
}

type RollappState struct {
	// BlockDescriptor is the block descriptor for the required height
	BlockDescriptor rollapptypes.BlockDescriptor
	// NextBlockSequencer is the tendermint pubkey of the sequencer who submitted the block descriptor for the next height (h+1)
	NextBlockSequencer tmprotocrypto.PublicKey
}
