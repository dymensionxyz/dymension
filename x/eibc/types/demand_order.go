package types

import (
	"crypto/sha256"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func NewDemandOrder(packetKey string, price string, fee string, denom string, recipient string) *DemandOrder {
	return &DemandOrder{
		Id:                   BuildDemandIDFromPacketKey(packetKey),
		TrackingPacketKey:    packetKey,
		Price:                price,
		Fee:                  fee,
		Denom:                denom,
		Recipient:            recipient,
		IsFullfilled:         false,
		TrackingPacketStatus: commontypes.Status_PENDING,
	}
}

func (m *DemandOrder) ValidateBasic() error {
	price, ok := sdk.NewIntFromString(m.Price)
	if !ok {
		return ErrInvalidAmount
	}
	if !price.IsPositive() {
		return ErrInvalidDemandOrderPrice
	}
	fee, ok := sdk.NewIntFromString(m.Fee)
	if !ok {
		return ErrInvalidAmount
	}
	if !fee.IsPositive() {
		return ErrInvalidDemandOrderFee
	}
	_, err := sdk.AccAddressFromBech32(m.Recipient)
	if err != nil {
		return ErrInvalidRecipientAddress
	}
	return ibctransfertypes.ValidatePrefixedDenom(m.Denom)
}

func (m *DemandOrder) Validate() error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

// GetPriceMathInt returns the price as a math.Int. Should
// be called after ValidateBasic hence should not panic.
func (m *DemandOrder) GetPriceInCoins() sdk.Coin {
	price, ok := sdk.NewIntFromString(m.Price)
	if !ok {
		panic("invalid price")
	}
	return sdk.NewCoin(m.Denom, price)
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
