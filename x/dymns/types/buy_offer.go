package types

import (
	"fmt"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// HasCounterpartyOfferPrice returns true if the offer has a raise-offer request from the Dym-Name owner.
func (m *BuyOrder) HasCounterpartyOfferPrice() bool {
	return m.CounterpartyOfferPrice != nil && !m.CounterpartyOfferPrice.Amount.IsNil() && !m.CounterpartyOfferPrice.IsZero()
}

// Validate performs basic validation for the BuyOrder.
func (m *BuyOrder) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "offer is nil")
	}

	if m.Id == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "ID of offer is empty")
	}

	if !IsValidBuyOrderId(m.Id) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "ID of offer is not a valid offer id")
	}

	switch m.AssetType {
	case TypeName:
		if !strings.HasPrefix(m.Id, BuyOrderIdTypeDymNamePrefix) {
			return errorsmod.Wrap(
				gerrc.ErrInvalidArgument,
				"mismatch type of Buy-Order ID prefix and type",
			)
		}

		if m.AssetId == "" {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name of offer is empty")
		}

		if !dymnsutils.IsValidDymName(m.AssetId) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name of offer is not a valid dym name")
		}
	case TypeAlias:
		if !strings.HasPrefix(m.Id, BuyOrderIdTypeAliasPrefix) {
			return errorsmod.Wrap(
				gerrc.ErrInvalidArgument,
				"mismatch type of Buy-Order ID prefix and type",
			)
		}

		if m.AssetId == "" {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias of offer is empty")
		}

		if !dymnsutils.IsValidAlias(m.AssetId) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias of offer is not a valid alias")
		}
	default:
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid asset type: %s", m.AssetType)
	}

	if err := ValidateOrderParams(m.Params, m.AssetType); err != nil {
		return err
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

// GetSdkEvent returns the sdk event contains information of BuyOrder record.
// Fired when BuyOrder record is set into store.
func (m BuyOrder) GetSdkEvent(actionName string) sdk.Event {
	var attrCounterpartyOfferPrice sdk.Attribute
	if m.CounterpartyOfferPrice != nil {
		attrCounterpartyOfferPrice = sdk.NewAttribute(AttributeKeyBoCounterpartyOfferPrice, m.CounterpartyOfferPrice.String())
	} else {
		attrCounterpartyOfferPrice = sdk.NewAttribute(AttributeKeyBoCounterpartyOfferPrice, "")
	}

	return sdk.NewEvent(
		EventTypeBuyOrder,
		sdk.NewAttribute(AttributeKeyBoId, m.Id),
		sdk.NewAttribute(AttributeKeyBoAssetId, m.AssetId),
		sdk.NewAttribute(AttributeKeyBoAssetType, m.AssetType.FriendlyString()),
		sdk.NewAttribute(AttributeKeyBoBuyer, m.Buyer),
		sdk.NewAttribute(AttributeKeyBoOfferPrice, m.OfferPrice.String()),
		attrCounterpartyOfferPrice,
		sdk.NewAttribute(AttributeKeyBoActionName, actionName),
	)
}

// IsValidBuyOrderId returns true if the given string is a valid ID for a Buy-Order record.
func IsValidBuyOrderId(id string) bool {
	if len(id) < 3 {
		return false
	}
	switch id[:2] {
	case BuyOrderIdTypeDymNamePrefix:
	case BuyOrderIdTypeAliasPrefix:
	default:
		return false
	}

	ui, err := strconv.ParseUint(id[2:], 10, 64)
	return err == nil && ui > 0
}

// CreateBuyOrderId creates a new BuyOrder ID from the given parameters.
func CreateBuyOrderId(_type AssetType, i uint64) string {
	var prefix string
	switch _type {
	case TypeName:
		prefix = BuyOrderIdTypeDymNamePrefix
	case TypeAlias:
		prefix = BuyOrderIdTypeAliasPrefix
	default:
		panic(fmt.Sprintf("unknown buy asset type: %d", _type))
	}

	buyOrderId := prefix + sdkmath.NewIntFromUint64(i).String()

	if !IsValidBuyOrderId(buyOrderId) {
		panic("bad input parameters for creating buy order id")
	}

	return buyOrderId
}
