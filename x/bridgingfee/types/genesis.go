package types

import "fmt"

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
			return fmt.Errorf("duplicate id: %d", id)
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
			return fmt.Errorf("duplicate id: %d", id)
		}
		seenIds[id] = true

		if err := hook.Validate(); err != nil {
			return err
		}
	}

	return nil
}
