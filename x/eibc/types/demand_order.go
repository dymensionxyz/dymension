package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// NewDemandOrder creates a new demand order.
// TODO: Change to rollapp packet
func NewDemandOrder(rollappPacket commontypes.RollappPacket, price string, fee string, denom string, recipient string) (*DemandOrder, error) {
	priceInt, ok := sdk.NewIntFromString(price)
	if !ok {
		return nil, ErrInvalidDemandOrderPrice
	}
	feeInt, ok := sdk.NewIntFromString(fee)
	if !ok {
		return nil, ErrInvalidDemandOrderFee
	}
	rollappPacketKey, err := commontypes.RollappPacketKey(&rollappPacket)
	if err != nil {
		return nil, err
	}

	return &DemandOrder{
		Id:                   BuildDemandIDFromPacketKey(string(rollappPacketKey)),
		TrackingPacketKey:    string(rollappPacketKey),
		Price:                sdk.NewCoins(sdk.NewCoin(denom, priceInt)),
		Fee:                  sdk.NewCoins(sdk.NewCoin(denom, feeInt)),
		Recipient:            recipient,
		IsFullfilled:         false,
		TrackingPacketStatus: commontypes.Status_PENDING,
	}, nil
}

func (m *DemandOrder) ValidateBasic() error {
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
	// Validate all tokens has a valid ibc denom
	for _, coin := range m.Price {
		if err := ibctransfertypes.ValidatePrefixedDenom(coin.Denom); err != nil {
			return err
		}
	}
	for _, coin := range m.Fee {
		if err := ibctransfertypes.ValidatePrefixedDenom(coin.Denom); err != nil {
			return err
		}
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
		sdk.NewAttribute(AttributeKeyIsFullfilled, strconv.FormatBool(m.IsFullfilled)),
		sdk.NewAttribute(AttributeKeyPacketStatus, m.TrackingPacketStatus.String()),
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

// BuildDemandIDFromPacketKey returns a unique demand order id from the packet key.
// PacketKey is used as a foreign key of rollapp packet in the demand order and as the demand order id.
// This is useful for when we want to get the demand order related to a specific rollapp packet and avoid
// from introducing another key for the demand order and double the storage.
func BuildDemandIDFromPacketKey(packetKey string) string {
	hash := sha256.Sum256([]byte(packetKey))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
