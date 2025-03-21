package types

// DefaultGenesisState returns the default genesis state for the hyperlane module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: *DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// TODO: Add validation logic
	return nil
}
