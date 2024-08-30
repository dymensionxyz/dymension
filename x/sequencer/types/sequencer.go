package types

import (
	cometbfttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// ValidateBasic performs basic validation of the sequencer object
func (seq Sequencer) ValidateBasic() error {
	if seq.Status == Unbonding && (seq.UnbondRequestHeight == 0 || seq.UnbondTime.IsZero()) {
		return ErrInvalidSequencerStatus
	}

	// validate notice period
	if seq.IsNoticePeriodInProgress() && seq.NoticePeriodTime.IsZero() {
		return ErrInvalidSequencerStatus
	}

	return nil
}

func (seq Sequencer) IsEmpty() bool {
	return seq.Address == ""
}

func (seq Sequencer) IsBonded() bool {
	return seq.Status == Bonded
}

// IsNoticePeriodInProgress returns true if the sequencer is bonded and has an unbond request
func (seq Sequencer) IsNoticePeriodInProgress() bool {
	return seq.Status == Bonded && seq.UnbondRequestHeight != 0
}

// GetDymintPubKeyHash returns the hash of the sequencer
// as expected to be written on the rollapp ibc client headers
func (seq Sequencer) GetDymintPubKeyHash() ([]byte, error) {
	// load the dymint pubkey into a cryptotypes.PubKey
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	var pubKey cryptotypes.PubKey
	err := protoCodec.UnpackAny(seq.DymintPubKey, &pubKey)
	if err != nil {
		return nil, err
	}

	// convert the pubkey to tmPubKey
	tmPubKey, err := cryptocodec.ToTmPubKeyInterface(pubKey)
	if err != nil {
		return nil, err
	}
	// Create a new tmValidator with fixed voting power of 1
	// TODO: Make sure the voting power is a param coming from hub and
	// not being set independently in dymint and hub
	tmValidator := cometbfttypes.NewValidator(tmPubKey, 1)
	tmValidatorSet := cometbfttypes.NewValidatorSet([]*cometbfttypes.Validator{tmValidator})
	return tmValidatorSet.Hash(), nil
}
