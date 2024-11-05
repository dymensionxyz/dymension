package types

import (
	"bytes"
	"errors"

	errorsmod "cosmossdk.io/errors"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// CheckCompatibility checks if the IBC state and Rollapp state are compatible
// Compatibility Criteria:
// 1. The app root shared by the IBC consensus state matches the block descriptor state root for the same height
// 2. The next validator hash shared by the IBC consensus state matches the sequencer hash for the next block descriptor
// 3. The block descriptor timestamp matches the tendermint header timestamp (only if timestamp exists for the block descriptor)
func CheckCompatibility(ibcState ibctm.ConsensusState, raState RollappState) error {
	// Check if block descriptor state root matches IBC block header app hash
	if !bytes.Equal(ibcState.Root.GetHash(), raState.BlockDescriptor.StateRoot) {
		return errorsmod.Wrap(ErrStateRootsMismatch, "block descriptor state root does not match tendermint header app hash")
	}
	// Check if the nextValidatorHash matches for the sequencer for h+1 block descriptor
	hash, err := raState.NextBlockSequencer.ValsetHash()
	if err != nil {
		return errors.Join(err, gerrc.ErrInternal.Wrap("val set hash"))
	}
	if !bytes.Equal(ibcState.NextValidatorsHash, hash) {
		return errorsmod.Wrap(ErrValidatorHashMismatch, "cons state next validator hash does not match the state info hash for sequencer for h+1")
	}
	if !raState.BlockDescriptor.Timestamp.IsZero() && !ibcState.Timestamp.Equal(raState.BlockDescriptor.Timestamp) {
		return errorsmod.Wrap(ErrTimestampMismatch, "block descriptor timestamp does not match tendermint header timestamp")
	}
	return nil
}

type RollappState struct {
	BlockDescriptor    rollapptypes.BlockDescriptor
	NextBlockSequencer sequencertypes.Sequencer
}
