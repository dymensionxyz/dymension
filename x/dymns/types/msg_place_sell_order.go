package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgPlaceSellOrder{}

// ValidateBasic performs basic validation for the MsgPlaceSellOrder.
func (m *MsgPlaceSellOrder) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	so := m.ToSellOrder()

	// put a dummy expire at to validate, as zero expire at is invalid,
	// and we don't have context of time at this point
	so.ExpireAt = 1

	if err := so.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order: %v", err)
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	return nil
}

// ToSellOrder converts the MsgPlaceSellOrder to a SellOrder.
func (m *MsgPlaceSellOrder) ToSellOrder() SellOrder {
	so := SellOrder{
		Name:      m.Name,
		MinPrice:  m.MinPrice,
		SellPrice: m.SellPrice,
	}

	if !so.HasSetSellPrice() {
		so.SellPrice = nil
	}

	return so
}

// GetSigners returns the required signers for the MsgPlaceSellOrder.
func (m *MsgPlaceSellOrder) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgPlaceSellOrder.
func (m *MsgPlaceSellOrder) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgPlaceSellOrder.
func (m *MsgPlaceSellOrder) Type() string {
	return TypeMsgPlaceSellOrder
}

// GetSignBytes returns the raw bytes for the MsgPlaceSellOrder.
func (m *MsgPlaceSellOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
