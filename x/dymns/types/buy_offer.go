package types

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"strconv"
	"strings"
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

	if !IsValidBuyOfferId(m.Id) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "ID of offer is not a valid offer id")
	}

	switch m.Type {
	case MarketOrderType_MOT_DYM_NAME:
		if !strings.HasPrefix(m.Id, BuyOfferIdTypeDymNamePrefix) {
			return errorsmod.Wrap(
				gerrc.ErrInvalidArgument,
				"mismatch type of Buy-Order ID prefix and type",
			)
		}
	case MarketOrderType_MOT_ALIAS:
		if !strings.HasPrefix(m.Id, BuyOfferIdTypeAliasPrefix) {
			return errorsmod.Wrap(
				gerrc.ErrInvalidArgument,
				"mismatch type of Buy-Order ID prefix and type",
			)
		}
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
		sdk.NewAttribute(AttributeKeyBoType, m.Type.String()),
		sdk.NewAttribute(AttributeKeyBoBuyer, m.Buyer),
		sdk.NewAttribute(AttributeKeyBoOfferPrice, m.OfferPrice.String()),
		attrCounterpartyOfferPrice,
		sdk.NewAttribute(AttributeKeyBoActionName, actionName),
	)
}

// IsValidBuyOfferId returns true if the given string is a valid offer-id for buy offer.
func IsValidBuyOfferId(id string) bool {
	if len(id) < 3 {
		return false
	}
	switch id[:2] {
	case BuyOfferIdTypeDymNamePrefix:
	case BuyOfferIdTypeAliasPrefix:
	default:
		return false
	}

	ui, err := strconv.ParseUint(id[2:], 10, 64)
	return err == nil && ui > 0
}

// CreateBuyOfferId creates a new BuyOffer ID from the given parameters.
func CreateBuyOfferId(_type MarketOrderType, i uint64) string {
	var prefix string
	switch _type {
	case MarketOrderType_MOT_DYM_NAME:
		prefix = BuyOfferIdTypeDymNamePrefix
	case MarketOrderType_MOT_ALIAS:
		prefix = BuyOfferIdTypeAliasPrefix
	default:
		panic(fmt.Sprintf("unknown buy offer type: %d", _type))
	}

	offerId := prefix + sdkmath.NewIntFromUint64(i).String()

	if !IsValidBuyOfferId(offerId) {
		panic("bad input parameters for creating buy offer id")
	}

	return offerId
}
