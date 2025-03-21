package types

// GenesisState defines the hyperlane module's genesis state.
type GenesisState struct {
	// TODO: Add genesis state fields
	Params *Params `json:"params" yaml:"params"`
}

// DefaultGenesisState returns the default genesis state for the hyperlane module.
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// TODO: Add validation logic
	return nil
}
