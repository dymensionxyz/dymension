package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

func (genState GenesisState) Validate() error {
	return nil
}
