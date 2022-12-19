package types

import (
	crypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ce "github.com/tendermint/tendermint/crypto/encoding"
)

const TypeMsgCreateSequencer = "create_sequencer"

var (
	_ sdk.Msg = &MsgCreateSequencer{}
)

func NewMsgCreateSequencer(creator string, sequencerAddress string, pubkey crypto.PublicKey, rollappId string, description *Description) (*MsgCreateSequencer, error) {
	return &MsgCreateSequencer{
		Creator:          creator,
		SequencerAddress: sequencerAddress,
		DymintPubKey:     pubkey,
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

	_, err = ce.PubKeyFromProto(msg.DymintPubKey)
	if err != nil {
		return err
	}

	// get Bech32 format
	_, err = sdk.AccAddressFromBech32(msg.SequencerAddress)
	if err != nil {
		return sdkerrors.Wrapf(ErrInvalidSequencerAddress, "invalid permissioned address: %s", err)
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return err
	}

	return nil
}
