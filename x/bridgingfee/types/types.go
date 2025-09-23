package types

import (
	"fmt"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
