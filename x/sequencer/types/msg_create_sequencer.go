package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/decred/dcrd/dcrec/edwards"
)

const (
	TypeMsgCreateSequencer = "create_sequencer"
	TypeMsgUnbond          = "unbond"
	TypeMsgIncreaseBond    = "increase_bond"
)

var (
	_ sdk.Msg                            = &MsgCreateSequencer{}
	_ sdk.Msg                            = &MsgUnbond{}
	_ sdk.Msg                            = &MsgIncreaseBond{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateSequencer)(nil)
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateSequencer) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.DymintPubKey, &pubKey)
}

/* --------------------------- MsgCreateSequencer --------------------------- */
func NewMsgCreateSequencer(creator string, pubkey cryptotypes.PubKey, rollappId string, metadata *SequencerMetadata, bond sdk.Coin) (*MsgCreateSequencer, error) {
	if metadata == nil {
		return nil, ErrInvalidRequest
	}
	var pkAny *codectypes.Any
	if pubkey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubkey); err != nil {
			return nil, err
		}
	}

	return &MsgCreateSequencer{
		Creator:      creator,
		DymintPubKey: pkAny,
		RollappId:    rollappId,
		Metadata:     *metadata,
		Bond:         bond,
	}, nil
}

func (msg *MsgCreateSequencer) Route() string {
	return RouterKey
}

func (msg *MsgCreateSequencer) Type() string {
	return TypeMsgCreateSequencer
}

func (msg *MsgCreateSequencer) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgCreateSequencer) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgCreateSequencer) ValidateBasic() error {
	return msg.Validate(false)
}

func (msg *MsgCreateSequencer) Validate(isEVM bool) error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// public key also checked by the application logic
	if msg.DymintPubKey == nil {
		return errorsmod.Wrap(ErrInvalidPubKey, "sequencer pubkey is required")
	}

	// check it is a pubkey
	if _, err = codectypes.NewAnyWithValue(msg.DymintPubKey); err != nil {
		return errorsmod.Wrapf(ErrInvalidPubKey, "invalid sequencer pubkey(%s)", err)
	}

	// cast to cryptotypes.PubKey type
	pk, ok := msg.DymintPubKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return errorsmod.Wrapf(ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk)
	}

	_, err = edwards.ParsePubKey(edwards.Edwards(), pk.Bytes())
	// err means the pubkey validation failed
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidPubKey, "%s", err)
	}

	if _, err = msg.Metadata.UpdateSequencerMetadata(msg.Metadata); err != nil {
		return err
	}

	if err = msg.Metadata.Validate(isEVM); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	if !msg.Bond.IsValid() {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.Bond.String())
	}

	return nil
}

/* ---------------------------- MsgIncreaseBond ---------------------------- */
func NewMsgIncreaseBond(creator string, addAmount sdk.Coin) *MsgIncreaseBond {
	return &MsgIncreaseBond{
		Creator:   creator,
		AddAmount: addAmount,
	}
}

func (msg *MsgIncreaseBond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if !(msg.AddAmount.IsValid() && msg.AddAmount.IsPositive()) {
		return errorsmod.Wrapf(ErrInvalidCoins, "invalid bond amount: %s", msg.AddAmount.String())
	}

	return nil
}

func (msg *MsgIncreaseBond) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
