package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (seq Sequencer) IsBonded() bool {
	if seq.Status != Bonded && seq.Status != Proposer {
		return false
	}
	return true
}

// GetDymintPubKeyHash returns the hash of the sequencer
// as expected to written on the rollapp ibc client headers
func (seq Sequencer) GetDymintPubKeyHash() ([]byte, error) {
	// load the dymint pubkey into a cryptotypes.PubKey
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
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
	tmValidator := tmtypes.NewValidator(tmPubKey, 1)
	tmValidatorSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{tmValidator})
	return tmValidatorSet.Hash(), nil
}
