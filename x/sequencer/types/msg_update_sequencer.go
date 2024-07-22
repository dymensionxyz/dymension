package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	TypeMsgUpdateSequencerInformation = "update_sequencer"
)

var (
	_ sdk.Msg                            = &MsgUpdateSequencerInformation{}
	_ sdk.Msg                            = &MsgUnbond{}
	_ codectypes.UnpackInterfacesMessage = (*MsgUpdateSequencerInformation)(nil)
)

func NewMsgUpdateSequencerInformation(creator string, rollappId string, metadata SequencerMetadata) (*MsgUpdateSequencerInformation, error) {
	return &MsgUpdateSequencerInformation{
		Creator:   creator,
		RollappId: rollappId,
		Metadata:  metadata,
	}, nil
}

func (msg *MsgUpdateSequencerInformation) Route() string {
	return RouterKey
}

func (msg *MsgUpdateSequencerInformation) Type() string {
	return TypeMsgUpdateSequencerInformation
}

func (msg *MsgUpdateSequencerInformation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateSequencerInformation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateSequencerInformation) ValidateBasic() error {
	if _, err := msg.Metadata.UpdateSequencerMetadata(msg.Metadata); err != nil {
		return err
	}

	return nil
}

func (msg *MsgUpdateSequencerInformation) UnpackInterfaces(codectypes.AnyUnpacker) error {
	return nil
}
