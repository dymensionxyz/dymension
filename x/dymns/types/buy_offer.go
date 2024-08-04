package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// HasCounterpartyOfferPrice returns true if the offer has a raise-offer request from the Dym-Name owner.
func (m *BuyOffer) HasCounterpartyOfferPrice() bool {
	return m.CounterpartyOfferPrice != nil && !m.CounterpartyOfferPrice.Amount.IsNil() && !m.CounterpartyOfferPrice.IsZero()
}

// Validate performs basic validation for the BuyOffer.
func (m *BuyOffer) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer is nil")
	}

	if m.Id == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "ID of offer is empty")
	}

	if !dymnsutils.IsValidBuyOfferId(m.Id) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "ID of offer is not a valid offer id")
	}

	if m.Name == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name of offer is empty")
	}

	if !dymnsutils.IsValidDymName(m.Name) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name of offer is not a valid dym name")
	}

	if m.Type != MarketOrderType_MOT_DYM_NAME {
		return errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"Buy-Order type must be: %s", MarketOrderType_MOT_DYM_NAME.String(),
		)
	}

	if !dymnsutils.IsValidBech32AccountAddress(m.Buyer, true) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "buyer is not a valid bech32 account address")
	}

	if m.OfferPrice.Amount.IsNil() || m.OfferPrice.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer price is zero")
	} else if m.OfferPrice.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer price is negative")
	} else if err := m.OfferPrice.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "offer price is invalid: %v", err)
	}

	if m.HasCounterpartyOfferPrice() {
		if m.CounterpartyOfferPrice.IsNegative() {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "counterparty offer price is negative")
		} else if err := m.CounterpartyOfferPrice.Validate(); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "counterparty offer price is invalid: %v", err)
		}

		if m.CounterpartyOfferPrice.Denom != m.OfferPrice.Denom {
			return errorsmod.Wrap(
				gerrc.ErrInvalidArgument,
				"counterparty offer price denom is different from offer price denom",
			)
		}
	}

	return nil
}

// GetSdkEvent returns the sdk event contains information of BuyOffer record.
// Fired when BuyOffer record is set into store.
func (m BuyOffer) GetSdkEvent(actionName string) sdk.Event {
	var attrCounterpartyOfferPrice sdk.Attribute
	if m.CounterpartyOfferPrice != nil {
		attrCounterpartyOfferPrice = sdk.NewAttribute(AttributeKeyBoCounterpartyOfferPrice, m.CounterpartyOfferPrice.String())
	} else {
		attrCounterpartyOfferPrice = sdk.NewAttribute(AttributeKeyBoCounterpartyOfferPrice, "")
	}

	return sdk.NewEvent(
		EventTypeBuyOffer,
		sdk.NewAttribute(AttributeKeyBoId, m.Id),
		sdk.NewAttribute(AttributeKeyBoName, m.Name),
		sdk.NewAttribute(AttributeKeyBoOfferPrice, m.OfferPrice.String()),
		attrCounterpartyOfferPrice,
		sdk.NewAttribute(AttributeKeyBoActionName, actionName),
	)
}
