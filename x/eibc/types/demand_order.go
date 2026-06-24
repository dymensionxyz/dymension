package types

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// NewDemandOrder creates a new demand order.
// Price is the cost to a market maker to buy the option, (recipient receives straight away).
// Fee is what the market maker gets in return.
func NewDemandOrder(rollappPacket commontypes.RollappPacket, price, fee math.Int, denom, recipient string, creationHeight uint64, completionHook *commontypes.CompletionHookCall, feeEscalation *FeeEscalation) *DemandOrder {
	rollappPacketKey := rollappPacket.RollappPacketKey()
	return &DemandOrder{
		Id:                   BuildDemandIDFromPacketKey(string(rollappPacketKey)),
		TrackingPacketKey:    string(rollappPacketKey),
		Price:                sdk.NewCoins(sdk.NewCoin(denom, price)),
		Fee:                  sdk.NewCoins(sdk.NewCoin(denom, fee)),
		Recipient:            recipient,
		TrackingPacketStatus: commontypes.Status_PENDING,
		RollappId:            rollappPacket.RollappId,
		Type:                 rollappPacket.Type,
		CreationHeight:       creationHeight,
		CompletionHook:       completionHook,
		FeeEscalation:        feeEscalation,
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
		return errors.Join(ErrInvalidRecipientAddress, err)
	}

	if m.CreationHeight == 0 {
		return ErrInvalidCreationHeight
	}

	return nil
}

func (m *DemandOrder) Validate() error {
	if err := m.ValidateBasic(); err != nil {
		return err
	}
	return nil
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

// PriceAmount returns the price amount of the demand order.
func (m *DemandOrder) PriceAmount() math.Int {
	return m.Price[0].Amount
}

// GetFeeAmount returns the fee amount of the demand order.
func (m *DemandOrder) GetFeeAmount() math.Int {
	return m.Fee.AmountOf(m.Price[0].Denom)
}

func (m *DemandOrder) GetFeePercent() math.LegacyDec {
	return math.LegacyNewDecFromInt(m.GetFeeAmount()).Quo(math.LegacyNewDecFromInt(m.PriceAmount()))
}

// escalates reports whether the order has an active, well-formed escalation spec
// and the given height is past creation. When false, all effective values equal
// the base values.
func (m *DemandOrder) escalates(height uint64) bool {
	return m.FeeEscalation != nil && m.FeeEscalation.DurationBlocks != 0 && height > m.CreationHeight
}

// EffectiveFeeAmount returns the fee offered at the given block height. It rises
// linearly from the base fee at creation to FeeEscalation.MaxFeeAmount after
// DurationBlocks, then stays flat. Equals GetFeeAmount() when not escalating.
func (m *DemandOrder) EffectiveFeeAmount(height uint64) math.Int {
	base := m.GetFeeAmount()
	if !m.escalates(height) {
		return base
	}
	elapsed := height - m.CreationHeight
	d := m.FeeEscalation.DurationBlocks
	if elapsed > d {
		elapsed = d
	}
	delta := m.FeeEscalation.MaxFeeAmount.Sub(base)
	inc := delta.Mul(math.NewIntFromUint64(elapsed)).Quo(math.NewIntFromUint64(d))
	return base.Add(inc)
}

// EffectivePriceAmount returns the price at the given block height. Escalating
// the fee is equivalent to lowering the price by the same amount. Equals
// PriceAmount() when not escalating.
func (m *DemandOrder) EffectivePriceAmount(height uint64) math.Int {
	base := m.PriceAmount()
	if !m.escalates(height) {
		return base
	}
	feeIncrease := m.EffectiveFeeAmount(height).Sub(m.GetFeeAmount())
	return base.Sub(feeIncrease)
}

// EffectiveFeePercent returns the fee/price ratio at the given block height.
func (m *DemandOrder) EffectiveFeePercent(height uint64) math.LegacyDec {
	return math.LegacyNewDecFromInt(m.EffectiveFeeAmount(height)).Quo(math.LegacyNewDecFromInt(m.EffectivePriceAmount(height)))
}

// ApplyEffectiveFee locks the order's Fee/Price to their height-effective values
// and clears the escalation spec, so the persisted record is self-consistent.
// No-op for non-escalating orders.
func (m *DemandOrder) ApplyEffectiveFee(height uint64) {
	if m.FeeEscalation == nil {
		return
	}
	denom := m.Denom()
	fee := m.EffectiveFeeAmount(height)
	price := m.EffectivePriceAmount(height)
	m.Fee = sdk.NewCoins(sdk.NewCoin(denom, fee))
	m.Price = sdk.NewCoins(sdk.NewCoin(denom, price))
	m.FeeEscalation = nil
}

func (m *DemandOrder) ValidateOrderIsOutstanding() error {
	// Check that the order is not fulfilled yet
	if m.IsFulfilled() {
		return ErrDemandAlreadyFulfilled
	}
	// Check the underlying packet is still relevant (i.e not expired, rejected, reverted)
	if m.TrackingPacketStatus != commontypes.Status_PENDING { // TODO:remove, there is only one callsite and it already knows it's pending
		return ErrDemandOrderInactive
	}
	return nil
}

func (m *DemandOrder) IsFulfilled() bool {
	return m.FulfillerAddress != "" || m.DeprecatedIsFulfilled
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

func (m *DemandOrder) Denom() string {
	// it's guaranteed price/fee are exactly one coin with the same denom
	return m.Price[0].Denom
}
