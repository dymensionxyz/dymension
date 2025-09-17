package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Validate validates the fee hook
func (h HLFeeHook) Validate() error {
	if h.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(h.Owner); err != nil {
			return fmt.Errorf("invalid owner: %s", h.Owner)
		}
	}
	for _, fee := range h.Fees {
		if err := fee.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate validates the asset fee
func (f HLAssetFee) Validate() error {
	if f.TokenID == "" {
		return fmt.Errorf("token id cannot be empty")
	}
	if f.InboundFee.IsNegative() {
		return fmt.Errorf("inbound fee cannot be negative")
	}
	if f.OutboundFee.IsNegative() {
		return fmt.Errorf("outbound fee cannot be negative")
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
