package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgOfferBuyName{}

// ValidateBasic performs basic validation for the MsgOfferBuyName.
func (m *MsgOfferBuyName) ValidateBasic() error {
	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "name is not a valid dym name")
	}

	if _, err := sdk.AccAddressFromBech32(m.Buyer); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	if m.ContinueOfferId != "" {
		if !dymnsutils.IsValidBuyNameOfferId(m.ContinueOfferId) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "continue offer id is not a valid buy name offer id")
		}
	}

	if !m.Offer.IsValid() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid offer amount")
	} else if !m.Offer.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer amount must be positive")
	}

	return nil
}

// GetSigners returns the required signers for the MsgOfferBuyName.
func (m *MsgOfferBuyName) GetSigners() []sdk.AccAddress {
	buyer, err := sdk.AccAddressFromBech32(m.Buyer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{buyer}
}

// Route returns the message router key for the MsgOfferBuyName.
func (m *MsgOfferBuyName) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgOfferBuyName.
func (m *MsgOfferBuyName) Type() string {
	return TypeMsgOfferBuyName
}

// GetSignBytes returns the raw bytes for the MsgOfferBuyName.
func (m *MsgOfferBuyName) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
