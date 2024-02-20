package types

import (
	"encoding/json"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgNonAvailableBatch = "submit_nonavailable"

var _ sdk.Msg = &MsgNonAvailableBatch{}

func NewMsgSubmitNonAvailableBatch(creator string, rollappID string, slIndex uint64, daPath string, nonInclusionProof string) *MsgNonAvailableBatch {
	return &MsgNonAvailableBatch{
		Creator:           creator,
		RollappId:         rollappID,
		SlIndex:           slIndex,
		DAPath:            daPath,
		NonInclusionProof: nonInclusionProof,
	}
}

func (msg *MsgNonAvailableBatch) Route() string {
	return RouterKey
}

func (msg *MsgNonAvailableBatch) Type() string {
	return TypeMsgNonAvailableBatch
}

func (msg *MsgNonAvailableBatch) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgNonAvailableBatch) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgNonAvailableBatch) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// Validate the JSON-encoded proof data
	/*_, err = msg.DecodeNonInclusionProof()
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "failed decoding fraud proof: %s", err)
	}*/
	return nil
}

func (msg *MsgNonAvailableBatch) DecodeNonInclusionProof() (inclusion.NonInclusionProof, error) {
	// Decode the JSON-encoded data into your struct
	nip := NonInclusionProof{}
	nonInclusionProof := inclusion.NonInclusionProof{}

	err := json.Unmarshal([]byte(msg.NonInclusionProof), &nip)
	if err != nil {
		return nonInclusionProof, fmt.Errorf("failed decoding JSON: %s", err)
	}

	nonInclusionProof.RowProof = nip.GetRproofs()

	return nonInclusionProof, nil

}
