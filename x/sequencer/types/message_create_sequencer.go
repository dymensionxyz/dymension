package types

import (
	"bytes"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgCreateSequencer = "create_sequencer"

var (
	_ sdk.Msg                            = &MsgCreateSequencer{}
	_ codectypes.UnpackInterfacesMessage = (*MsgCreateSequencer)(nil)
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (msg MsgCreateSequencer) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	var pubKey cryptotypes.PubKey
	return unpacker.UnpackAny(msg.DymintPubKey, &pubKey)
}

func NewMsgCreateSequencer(creator string, sequencerAddress string, pubkey cryptotypes.PubKey, rollappId string, description *Description) (*MsgCreateSequencer, error) {
	var pkAny *codectypes.Any
	if pubkey != nil {
		var err error
		if pkAny, err = codectypes.NewAnyWithValue(pubkey); err != nil {
			return nil, err
		}
	}
	return &MsgCreateSequencer{
		Creator:          creator,
		SequencerAddress: sequencerAddress,
		DymintPubKey:     pkAny,
		RollappId:        rollappId,
		Description:      *description,
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
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// check Bech32 format
	if _, err := sdk.AccAddressFromBech32(msg.SequencerAddress); err != nil {
		return sdkerrors.Wrapf(ErrInvalidSequencerAddress, "invalid permissioned address: %s", err)
	}

	// public key also checked by the application logic
	if msg.DymintPubKey != nil {
		// check it is a pubkey
		if _, err = codectypes.NewAnyWithValue(msg.DymintPubKey); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "invalid sequencer pubkey(%s)", err)
		}

		// cast to cryptotypes.PubKey type
		pk, ok := msg.DymintPubKey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
		}

		// get Bech32 format
		sequencerAddress, err := sdk.AccAddressFromBech32(msg.SequencerAddress)
		if err != nil {
			return sdkerrors.Wrapf(ErrInvalidSequencerAddress, "invalid permissioned address: %s", err)
		}

		// verify pubkey match the address
		if !bytes.Equal(pk.Address().Bytes(), sequencerAddress.Bytes()) {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidPubKey, "account address and pubkey address do not match")
		}
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}

	return nil
}
