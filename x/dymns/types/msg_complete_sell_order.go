package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgCompleteSellOrder{}

// ValidateBasic performs basic validation for the MsgCompleteSellOrder.
func (m *MsgCompleteSellOrder) ValidateBasic() error {
	if m.AssetType == TypeName {
		if !dymnsutils.IsValidDymName(m.AssetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "name is not a valid dym name: %s", m.AssetId)
		}
	} else if m.AssetType == TypeAlias {
		if !dymnsutils.IsValidAlias(m.AssetId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias is not a valid alias: %s", m.AssetId)
		}
	} else {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", m.AssetType)
	}

	if _, err := sdk.AccAddressFromBech32(m.Participant); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "participant is not a valid bech32 account address")
	}

	return nil
}

// GetSigners returns the required signers for the MsgCompleteSellOrder.
func (m *MsgCompleteSellOrder) GetSigners() []sdk.AccAddress {
	participant, err := sdk.AccAddressFromBech32(m.Participant)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{participant}
}

// Route returns the message router key for the MsgCompleteSellOrder.
func (m *MsgCompleteSellOrder) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgCompleteSellOrder.
func (m *MsgCompleteSellOrder) Type() string {
	return TypeMsgCompleteSellOrder
}

// GetSignBytes returns the raw bytes for the MsgCompleteSellOrder.
func (m *MsgCompleteSellOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
