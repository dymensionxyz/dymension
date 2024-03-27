package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/decred/dcrd/dcrec/edwards"
)

const (
	TypeMsgCreateSequencer = "create_sequencer"
	TypeMsgUnbond          = "unbond"
)

var (
	_ sdk.Msg                            = &MsgCreateSequencer{}
	_ sdk.Msg                            = &MsgUnbond{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateSequencer)(nil)
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateSequencer) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.DymintPubKey, &pubKey)
}

/* --------------------------- MsgCreateSequencer --------------------------- */
func NewMsgCreateSequencer(creator string, pubkey cryptotypes.PubKey, rollappId string, description *Description, bond sdk.Coin) (*MsgCreateSequencer, error) {
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
		Description:  *description,
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
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// public key also checked by the application logic
	if msg.DymintPubKey != nil {
		// check it is a pubkey
		if _, err = codectypes.NewAnyWithValue(msg.DymintPubKey); err != nil {
			return errorsmod.Wrapf(errortypes.ErrInvalidPubKey, "invalid sequencer pubkey(%s)", err)
		}

		// cast to cryptotypes.PubKey type
		pk, ok := msg.DymintPubKey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			return errorsmod.Wrapf(errortypes.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
		}

		_, err = edwards.ParsePubKey(edwards.Edwards(), pk.Bytes())
		// err means the pubkey validation failed
		if err != nil {
			return errorsmod.Wrapf(errortypes.ErrInvalidPubKey, "%s", err)
		}

	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}

	if !msg.Bond.IsValid() {
		return errorsmod.Wrapf(errortypes.ErrInvalidCoins, "invalid bond amount: %s", msg.Bond.String())
	}

	return nil
}

/* -------------------------------- MsgUnbond ------------------------------- */
func NewMsgUnbond(creator string) *MsgUnbond {
	return &MsgUnbond{
		Creator: creator,
	}
}

func (msg *MsgUnbond) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return errorsmod.Wrapf(errortypes.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}

func (msg *MsgUnbond) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}
