package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		FeeHooks:         []HLFeeHook{},
		AggregationHooks: []AggregationHook{},
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// Validate fee hooks
	seenIds := make(map[uint64]bool)
	for _, hook := range gs.FeeHooks {
		id := hook.Id.GetInternalId()
		if seenIds[id] {
			return ErrDuplicateHookId
		}
		seenIds[id] = true

		if err := hook.Validate(); err != nil {
			return err
		}
	}

	// Validate aggregation hooks
	for _, hook := range gs.AggregationHooks {
		id := hook.Id.GetInternalId()
		if seenIds[id] {
			return ErrDuplicateHookId
		}
		seenIds[id] = true

		if err := hook.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the fee hook
func (h HLFeeHook) Validate() error {
	if h.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(h.Owner); err != nil {
			return err
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
		return ErrInvalidTokenId
	}

	if f.InboundFee.IsNegative() {
		return ErrInvalidFee
	}

	if f.OutboundFee.IsNegative() {
		return ErrInvalidFee
	}

	return nil
}

// Validate validates the aggregation hook
func (h AggregationHook) Validate() error {
	if h.Owner != "" {
		if _, err := sdk.AccAddressFromBech32(h.Owner); err != nil {
			return err
		}
	}

	if len(h.HookIds) == 0 {
		return ErrEmptyHookIds
	}

	return nil
}
