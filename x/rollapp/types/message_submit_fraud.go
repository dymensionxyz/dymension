package types

import (
	"encoding/json"
	fmt "fmt"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

const TypeMsgSubmitFraud = "submit_fraud"

var _ sdk.Msg = &MsgSubmitFraud{}

func NewMsgSubmitFraud(creator string, rollappID string, fraudproof string) *MsgSubmitFraud {
	return &MsgSubmitFraud{
		Creator:    creator,
		RollappID:  rollappID,
		FraudProof: fraudproof,
	}
}

func (msg *MsgSubmitFraud) Route() string {
	return RouterKey
}

func (msg *MsgSubmitFraud) Type() string {
	return TypeMsgSubmitFraud
}

func (msg *MsgSubmitFraud) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgSubmitFraud) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgSubmitFraud) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	// Validate the rollapp ID
	if len(msg.RollappID) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "rollapp ID cannot be empty")
	}

	// Validate the JSON-encoded fraudproof data
	_, err = msg.DecodeFraudProof()
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "failed decoding fraud proof: %s", err)
	}
	return nil
}

func (msg *MsgSubmitFraud) DecodeFraudProof() (fraudtypes.FraudProof, error) {
	// Decode the JSON-encoded data into your struct
	fp := abcitypes.FraudProof{}
	fraud := fraudtypes.FraudProof{}

	err := json.Unmarshal([]byte(msg.FraudProof), &fp)
	if err != nil {
		return fraud, fmt.Errorf("failed decoding JSON: %s", err)
	}

	err = fraud.FromABCI(fp)
	if err != nil {
		return fraud, fmt.Errorf("failed decoding JSON: %s", err)
	}

	return fraud, nil
}
