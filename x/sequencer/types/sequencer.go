package types

import (
	"slices"
	"time"

	errorsmod "cosmossdk.io/errors"
	comettypes "github.com/cometbft/cometbft/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

const (
	SentinelSeqAddr = "sentinel"
)

func NewTestSequencer(
	pk cryptotypes.PubKey,
) Sequencer {
	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		panic(err)
	}
	return Sequencer{
		Address:      pk.Address().String(),
		DymintPubKey: pkAny,
	}
}

// ValidateBasic performs basic validation of the sequencer object
func (seq Sequencer) ValidateBasic() error {
	if seq.Tokens.Len() != 1 {
		return gerrc.ErrInvalidArgument.Wrap("expect one coin")
	}
	return nil
}

func (seq *Sequencer) SetOptedIn(ctx sdk.Context, x bool) error {
	if err := uevent.EmitTypedEvent(ctx, &EventOptInStatusChange{
		seq.RollappId,
		seq.Address,
		seq.OptedIn,
		x,
	}); err != nil {
		return err
	}
	seq.OptedIn = x
	return nil
}

func (seq Sequencer) Sentinel() bool {
	return seq.Address == SentinelSeqAddr
}

func (seq Sequencer) Bonded() bool {
	return seq.Status == Bonded
}

func (seq Sequencer) IsPotentialProposer() bool {
	return seq.Bonded() && seq.OptedIn
}

func (seq Sequencer) TokensCoin() sdk.Coin {
	return seq.Tokens[0]
}

func (seq Sequencer) SetTokensCoin(c sdk.Coin) {
	seq.Tokens[0] = c
}

func (seq Sequencer) AccAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(seq.Address)
}

func (seq Sequencer) NoticeInProgress(now time.Time) bool {
	return seq.NoticeStarted() && !seq.NoticeElapsed(now)
}

func (seq Sequencer) NoticeElapsed(now time.Time) bool {
	return seq.NoticeStarted() && !now.Before(seq.NoticePeriodTime)
}

func (seq Sequencer) NoticeStarted() bool {
	return seq.NoticePeriodTime != time.Time{}
}

// Also called 'dymint proposer addr' in some places
func (seq Sequencer) ProposerAddr() ([]byte, error) {
	return PubKeyAddr(seq.DymintPubKey)
}

func (seq *Sequencer) SetWhitelistedRelayers(relayers []string) {
	slices.Sort(relayers)
	seq.WhitelistedRelayers = relayers
}

// MustProposerAddr : intended for tests
func (seq Sequencer) MustProposerAddr() []byte {
	ret, err := seq.ProposerAddr()
	if err != nil {
		panic(err)
	}
	return ret
}

// MustPubKey is intended for tests
func (seq Sequencer) MustPubKey() cryptotypes.PubKey {
	x, err := PubKey(seq.DymintPubKey)
	if err != nil {
		panic(err)
	}
	return x
}

func (seq Sequencer) Valset() (*comettypes.ValidatorSet, error) {
	pubKey, err := PubKey(seq.DymintPubKey)
	if err != nil {
		return nil, errorsmod.Wrap(err, "pub key")
	}
	return Valset(pubKey)
}

func (seq Sequencer) ValsetHash() ([]byte, error) {
	pubKey, err := PubKey(seq.DymintPubKey)
	if err != nil {
		return nil, errorsmod.Wrap(err, "pub key")
	}
	return ValsetHash(pubKey)
}

// MustValset : intended for tests
func (seq Sequencer) MustValset() *comettypes.ValidatorSet {
	x, err := seq.Valset()
	if err != nil {
		panic(err)
	}
	return x
}

// MustValsetHash : intended for tests
func (seq Sequencer) MustValsetHash() []byte {
	x, err := seq.ValsetHash()
	if err != nil {
		panic(err)
	}
	return x
}

var _ codectypes.UnpackInterfacesMessage = (*Sequencer)(nil)

func (s Sequencer) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(s.DymintPubKey, &pubKey)
}

// TODO: move these utils to a more suitable package

func PubKey(pk *codectypes.Any) (cryptotypes.PubKey, error) {
	cdc := ModuleCdc2
	var pubKey cryptotypes.PubKey
	err := cdc.UnpackAny(pk, &pubKey)
	return pubKey, err
}

// PubKeyAddr returns comet/dymint 'proposer address' if pkA is an ed25519 pubkey
func PubKeyAddr(pkA *codectypes.Any) ([]byte, error) {
	pk, err := PubKey(pkA)
	if err != nil {
		return nil, err
	}
	return pk.Address(), nil
}

func Valset(pubKey cryptotypes.PubKey) (*comettypes.ValidatorSet, error) {
	// convert the pubkey to tmPubKey
	tmPubKey, err := cryptocodec.ToTmPubKeyInterface(pubKey)
	if err != nil {
		return nil, errorsmod.Wrap(err, "tm pub key")
	}

	val := comettypes.NewValidator(tmPubKey, 1)

	return comettypes.ValidatorSetFromExistingValidators([]*comettypes.Validator{
		val,
	})
}

func ValsetHash(pubKey cryptotypes.PubKey) ([]byte, error) {
	vs, err := Valset(pubKey)
	if err != nil {
		return nil, err
	}
	return vs.Hash(), nil
}
