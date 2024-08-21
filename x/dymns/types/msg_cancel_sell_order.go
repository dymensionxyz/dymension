package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgCancelSellOrder{}

// ValidateBasic performs basic validation for the MsgCancelSellOrder.
func (m *MsgCancelSellOrder) ValidateBasic() error {
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

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	return nil
}

// GetSigners returns the required signers for the MsgCancelSellOrder.
func (m *MsgCancelSellOrder) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgCancelSellOrder.
func (m *MsgCancelSellOrder) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgCancelSellOrder.
func (m *MsgCancelSellOrder) Type() string {
	return TypeMsgCancelSellOrder
}

// GetSignBytes returns the raw bytes for the MsgCancelSellOrder.
func (m *MsgCancelSellOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
