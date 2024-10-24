package types

import (
	"time"

	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	cometbfttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	SentinelSeqAddr = "sentinel"
)

// ValidateBasic performs basic validation of the sequencer object
func (seq Sequencer) ValidateBasic() error {
	if seq.Tokens.Len() != 1 {
		return gerrc.ErrInvalidArgument.Wrap("expect one coin")
	}
	return nil
}

func (seq Sequencer) Sentinel() bool {
	return seq.Address == SentinelSeqAddr
}

func (seq Sequencer) Bonded() bool {
	return seq.Status == Bonded
}

func (seq Sequencer) TokensCoin() sdk.Coin {
	return seq.Tokens[0]
}

func (seq Sequencer) SetTokensCoin(c sdk.Coin) {
	seq.Tokens = sdk.Coins{c}
}

func (seq Sequencer) AccAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(seq.Address)
}

func (seq Sequencer) NoticeInProgress(now time.Time) bool {
	return seq.NoticeStarted() && !seq.NoticeElapsed(now)
}

func (seq Sequencer) NoticeElapsed(now time.Time) bool {
	return seq.NoticeStarted() && seq.NoticePeriodTime.Before(now)
}

func (seq Sequencer) NoticeStarted() bool {
	return seq.NoticePeriodTime != time.Time{}
}

// GetDymintPubKeyHash returns the hash of the sequencer
// as expected to be written on the rollapp ibc client headers
func (seq Sequencer) GetDymintPubKeyHash() ([]byte, error) {
	pubKey, err := seq.cosmosPubKey()
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

// CometPubKey returns the bytes of the sequencer's dymint pubkey
func (seq Sequencer) CometPubKey() (tmprotocrypto.PublicKey, error) {
	pubKey, err := seq.cosmosPubKey()
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}

	// convert the pubkey to tmPubKey
	tmPubKey, err := cryptocodec.ToTmProtoPublicKey(pubKey)
	return tmPubKey, err
}

func (seq Sequencer) cosmosPubKey() (cryptotypes.PubKey, error) {
	interfaceRegistry := cdctypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	var pubKey cryptotypes.PubKey
	err := protoCodec.UnpackAny(seq.DymintPubKey, &pubKey)
	return pubKey, err
}
