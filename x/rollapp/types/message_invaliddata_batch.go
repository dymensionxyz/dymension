package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgInvalidDataBatch = "submit_invaliddata"

var _ sdk.Msg = &MsgInvalidDataBatch{}

func NewMsgInvalidDataBatch(creator string, rollappID string, slIndex uint64, daPath string, inclusionProof string) *MsgWrongCommitmentBatch {

	return &MsgWrongCommitmentBatch{
		Creator:        creator,
		RollappId:      rollappID,
		SlIndex:        slIndex,
		DAPath:         daPath,
		InclusionProof: inclusionProof,
	}
}

func (msg *MsgInvalidDataBatch) Route() string {
	return RouterKey
}

func (msg *MsgInvalidDataBatch) Type() string {
	return TypeMsgNonAvailableBatch
}

func (msg *MsgInvalidDataBatch) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgInvalidDataBatch) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgInvalidDataBatch) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	return nil
}
