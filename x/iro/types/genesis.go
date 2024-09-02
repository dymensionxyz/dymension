package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {

	return gs.Params.Validate()

	// validate lastPlan ID (GT 0, sequential)

	// validate no multiple plans for the same rollapp

	// validate each plan
}
