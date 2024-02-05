package types

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	common "github.com/dymensionxyz/dymension/v3/x/common"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func NewDemandOrder(packetKey string, price string, fee string, denom string, recipient string) *DemandOrder {
	return &DemandOrder{
		Id:                   BuildDemandIDFromPacketKey(packetKey),
		TrackingPacketKey:    packetKey,
		Price:                sdk.NewCoins(sdk.NewCoin(denom, common.StringToSdkInt(price))),
		Fee:                  sdk.NewCoins(sdk.NewCoin(denom, common.StringToSdkInt(fee))),
		Recipient:            recipient,
		IsFullfilled:         false,
		TrackingPacketStatus: commontypes.Status_PENDING,
	}
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
		sdk.NewAttribute(AttributeKeyPacketKey, m.Id),
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

func BuildDemandIDFromPacketKey(packetKey string) string {
	hash := sha256.Sum256([]byte(packetKey))
	hashString := hex.EncodeToString(hash[:])
	return hashString
}
