package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgAcceptOfferBuyName{}

// ValidateBasic performs basic validation for the MsgAcceptOfferBuyName.
func (m *MsgAcceptOfferBuyName) ValidateBasic() error {
	if !dymnsutils.IsValidBuyOfferId(m.OfferId) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer id is not a valid buy name offer id")
	}

	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "owner is not a valid bech32 account address")
	}

	if !m.MinAccept.IsValid() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid min-accept amount")
	} else if !m.MinAccept.IsPositive() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "min-accept amount must be positive")
	}

	return nil
}

// GetSigners returns the required signers for the MsgAcceptOfferBuyName.
func (m *MsgAcceptOfferBuyName) GetSigners() []sdk.AccAddress {
	owner, err := sdk.AccAddressFromBech32(m.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{owner}
}

// Route returns the message router key for the MsgAcceptOfferBuyName.
func (m *MsgAcceptOfferBuyName) Route() string {
	return RouterKey
}

// Type returns the message type for the MsgAcceptOfferBuyName.
func (m *MsgAcceptOfferBuyName) Type() string {
	return TypeMsgAcceptOfferBuyName
}

// GetSignBytes returns the raw bytes for the MsgAcceptOfferBuyName.
func (m *MsgAcceptOfferBuyName) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(m)
	return sdk.MustSortJSON(bz)
}
