package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgPlaceBuyOrder{}

// ValidateBasic performs basic validation for the MsgPlaceBuyOrder.
func (m *MsgPlaceBuyOrder) ValidateBasic() error {
	if m.OrderType == NameOrder {
		if !dymnsutils.IsValidDymName(m.GoodsId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "name is not a valid dym name: %s", m.GoodsId)
		}
	} else if m.OrderType == AliasOrder {
		if !dymnsutils.IsValidAlias(m.GoodsId) {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "alias is not a valid alias: %s", m.GoodsId)
		}
	} else {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid order type: %s", m.OrderType)
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	if m.ContinueOfferId != "" && !IsValidBuyOfferId(m.ContinueOfferId) {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument,
			"continue offer id is not a valid offer id: %s", m.ContinueOfferId,
		)
	}

	if !m.Offer.IsValid() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid offer amount")
	} else if !m.Offer.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer amount must be positive")
	}

	return nil
}

// GetSigners returns the required signers for the MsgPlaceBuyOrder.
func (m *MsgPlaceBuyOrder) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

// Route returns the message router key for the MsgPlaceBuyOrder.
func (m *MsgPlaceBuyOrder) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgPlaceBuyOrder.
func (m *MsgPlaceBuyOrder) Type() string {
	return TypeMsgPlaceBuyOrder
}

// GetSignBytes returns the raw bytes for the MsgPlaceBuyOrder.
func (m *MsgPlaceBuyOrder) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
