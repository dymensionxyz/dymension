package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// NewDemandOrder creates a new demand order.
// Price is the cost to a market maker to buy the option, (recipient receives straight away).
// Fee is what the market maker gets in return.
func NewDemandOrder(rollappPacket commontypes.RollappPacket, price, fee math.Int, denom, recipient string) *DemandOrder {
	rollappPacketKey := commontypes.RollappPacketKey(&rollappPacket)
	return &DemandOrder{
		Id:                   BuildDemandIDFromPacketKey(string(rollappPacketKey)),
		TrackingPacketKey:    string(rollappPacketKey),
		Price:                sdk.NewCoins(sdk.NewCoin(denom, price)),
		Fee:                  sdk.NewCoins(sdk.NewCoin(denom, fee)),
		Recipient:            recipient,
		TrackingPacketStatus: commontypes.Status_PENDING,
		RollappId:            rollappPacket.RollappId,
		Type:                 rollappPacket.Type,
	}
}

func (m *DemandOrder) ValidateBasic() error {
	if len(m.Price) > 1 || len(m.Fee) > 1 {
		return ErrMultipleDenoms
	}

	if len(m.Price) == 0 {
		return ErrEmptyPrice
	}

	denom := m.Price[0].Denom

	// fee is optional, as it can be zero
	if len(m.Fee) != 0 && m.Fee[0].Denom != denom {
		return ErrMultipleDenoms
	}
	// Validate tokens has a valid ibc denom
	if err := ibctransfertypes.ValidatePrefixedDenom(denom); err != nil {
		return err
	}

	if err := m.Price.Validate(); err != nil {
		return err
	}
	if err := m.Fee.Validate(); err != nil {
		return err
	}
	_, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		return ErrInvalidRecipientAddress
	}

	// Validate the tracking packet key

	return nil
}

func (m *DemandOrder) Validate() error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

func (m *DemandOrder) GetEvents() []sdk.Attribute {
	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(AttributeKeyId, m.Id),
		sdk.NewAttribute(AttributeKeyPrice, m.Price.String()),
		sdk.NewAttribute(AttributeKeyFee, m.Fee.String()),
		sdk.NewAttribute(AttributeKeyIsFulfilled, strconv.FormatBool(m.IsFulfilled())),
		sdk.NewAttribute(AttributeKeyPacketStatus, m.TrackingPacketStatus.String()),
		sdk.NewAttribute(AttributeKeyPacketKey, m.TrackingPacketKey),
		sdk.NewAttribute(AttributeKeyRollappId, m.RollappId),
		sdk.NewAttribute(AttributeKeyRecipient, m.Recipient),
	}

	if m.FulfillerAddress != "" {
		eventAttributes = append(eventAttributes, sdk.NewAttribute(AttributeKeyFulfillerAddress, m.FulfillerAddress))
	}

	return eventAttributes
}

// GetRecipientBech32Address returns the recipient address as a string.
// Should be called after ValidateBasic hence should not panic.
func (m *DemandOrder) GetRecipientBech32Address() sdk.AccAddress {
	recipientBech32, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		panic(ErrInvalidRecipientAddress)
	}
	return recipientBech32
}

// GetFeeAmount returns the fee amount of the demand order.
func (m *DemandOrder) GetFeeAmount() math.Int {
	return m.Fee.AmountOf(m.Price[0].Denom)
}

func (m *DemandOrder) ValidateOrderIsOutstanding() error {
	// Check that the order is not fulfilled yet
	if m.IsFulfilled() {
		return ErrDemandAlreadyFulfilled
	}
	// Check the underlying packet is still relevant (i.e not expired, rejected, reverted)
	if m.TrackingPacketStatus != commontypes.Status_PENDING {
		return ErrDemandOrderInactive
	}
	return nil
}

func (m *DemandOrder) IsFulfilled() bool {
	return m.FulfillerAddress != ""
}

// BuildDemandIDFromPacketKey returns a unique demand order id from the packet key.
// PacketKey is used as a foreign key of rollapp packet in the demand order and as the demand order id.
// This is useful for when we want to get the demand order related to a specific rollapp packet and avoid
// from introducing another key for the demand order and double the storage.
func BuildDemandIDFromPacketKey(packetKey string) string {
	hash := sha256.Sum256([]byte(packetKey))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
