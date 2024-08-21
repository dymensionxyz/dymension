package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgPurchaseOrder{}

// ValidateBasic performs basic validation for the MsgPurchaseOrder.
func (m *MsgPurchaseOrder) ValidateBasic() error {
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

	if err := ValidateOrderParams(m.Params, m.AssetType); err != nil {
		return err
	}

	if !m.Offer.IsValid() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid offer")
	} else if !m.Offer.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer must be positive")
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	return nil
}

// GetSigners returns the required signers for the MsgPurchaseOrder.
func (m *MsgPurchaseOrder) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

// Route returns the message router key for the MsgPurchaseOrder.
func (m *MsgPurchaseOrder) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgPurchaseOrder.
func (m *MsgPurchaseOrder) Type() string {
	return TypeMsgPurchaseOrder
}

// GetSignBytes returns the raw bytes for the MsgPurchaseOrder.
func (m *MsgPurchaseOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
