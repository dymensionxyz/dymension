package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUpdateState = "update_state"

var _ sdk.Msg = &MsgUpdateState{}

func NewMsgUpdateState(creator string, rollappId string, startHeight uint64, numBlocks uint64, dAPath string, version uint64, lastBD *BlockDescriptor, bDs *BlockDescriptors) *MsgUpdateState {
	return &MsgUpdateState{
		Creator:     creator,
		RollappId:   rollappId,
		StartHeight: startHeight,
		NumBlocks:   numBlocks,
		DAPath:      dAPath,
		Version:     version,
		LastBD:      lastBD,
		BDs:         bDs,
	}
}

func (msg *MsgUpdateState) Route() string {
	return RouterKey
}

func (msg *MsgUpdateState) Type() string {
	return TypeMsgUpdateState
}

func (msg *MsgUpdateState) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateState) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateState) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
