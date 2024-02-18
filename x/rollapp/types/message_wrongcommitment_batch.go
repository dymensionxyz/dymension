package types

import (
	"encoding/json"
	fmt "fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
)

const TypeMsgWrongCommitmentBatch = "submit_wrongcommitment"

var _ sdk.Msg = &MsgWrongCommitmentBatch{}

func NewMsgWrongCommitmentBatch(creator string, rollappID string, slIndex uint64, daPath string, inclusionProof string) *MsgWrongCommitmentBatch {

	return &MsgWrongCommitmentBatch{
		Creator:        creator,
		RollappId:      rollappID,
		SlIndex:        slIndex,
		DAPath:         daPath,
		InclusionProof: inclusionProof,
	}
}

func (msg *MsgWrongCommitmentBatch) Route() string {
	return RouterKey
}

func (msg *MsgWrongCommitmentBatch) Type() string {
	return TypeMsgNonAvailableBatch
}

func (msg *MsgWrongCommitmentBatch) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgWrongCommitmentBatch) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgWrongCommitmentBatch) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	// Validate the JSON-encoded fraudproof data
	_, err = msg.DecodeInclusionProof()
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "failed decoding fraud proof: %s", err)
	}
	return nil
}

func (msg *MsgWrongCommitmentBatch) DecodeInclusionProof() (inclusion.InclusionProof, error) {
	// Decode the JSON-encoded data into your struct
	ip := BlobInclusionProof{}
	inclusionProof := inclusion.InclusionProof{}

	err := json.Unmarshal([]byte(msg.InclusionProof), &ip)
	if err != nil {
		return inclusionProof, fmt.Errorf("failed decoding JSON: %s", err)
	}

	inclusionProof.Blob = ip.GetBlob()
	inclusionProof.DataRoot = ip.GetDataroot()
	inclusionProof.Nmtproofs = ip.GetNmtproofs()
	inclusionProof.Nmtroots = ip.GetNmtroots()
	inclusionProof.RowProofs = ip.GetRproofs()

	return inclusionProof, nil

}
