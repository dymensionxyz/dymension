package types

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NormInt normalizes a possibly-nil math.Int to zero. Fields persisted before
// the min/max outbound fee bounds existed deserialize with an uninitialized
// (nil) big.Int; treating those as zero keeps old hooks behaving as before.
func NormInt(i math.Int) math.Int {
	if i.IsNil() {
		return math.ZeroInt()
	}
	return i
}

// Validate validates the fee hook
func (h HLFeeHook) Validate() error {
	if h.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(h.Owner); err != nil {
			return fmt.Errorf("invalid owner: %s", h.Owner)
		}
	}

	feeSet := make(map[util.HexAddress]struct{}, len(h.Fees))
	for _, fee := range h.Fees {
		if err := fee.Validate(); err != nil {
			return err
		}

		if _, ok := feeSet[fee.TokenId]; ok {
			return fmt.Errorf("duplicate fee entry: %s", fee.TokenId)
		}
		feeSet[fee.TokenId] = struct{}{}
	}
	return nil
}

// Validate validates the asset fee
func (f HLAssetFee) Validate() error {
	if f.TokenId.IsZeroAddress() {
		return fmt.Errorf("token id cannot be empty")
	}
	if f.InboundFee.IsNegative() {
		return fmt.Errorf("inbound fee cannot be negative")
	}
	if f.OutboundFee.IsNegative() {
		return fmt.Errorf("outbound fee cannot be negative")
	}
	if NormInt(f.MinOutboundFee).IsNegative() {
		return fmt.Errorf("min outbound fee cannot be negative")
	}
	if NormInt(f.MaxOutboundFee).IsNegative() {
		return fmt.Errorf("max outbound fee cannot be negative")
	}
	if mx := NormInt(f.MaxOutboundFee); mx.IsPositive() && mx.LT(NormInt(f.MinOutboundFee)) {
		return fmt.Errorf("max outbound fee must be >= min outbound fee")
	}
	return nil
}

// Validate validates the aggregation hook
func (h AggregationHook) Validate() error {
	if h.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(h.Owner); err != nil {
			return fmt.Errorf("owner address is invalid: %s", err.Error())
		}
	}
	return nil
}
