package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (g GenesisState) Validate() error {
	err := g.Params.Validate()
	if err != nil {
		return ErrInvalidGenesis.Wrapf(err.Error())
	}
	return nil
}
