package types

func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

func (g GenesisState) Validate() error {
	for _, fp := range g.RevokedPolicies {
		if err := ValidateFingerprint(fp); err != nil {
			return err
		}
	}
	return g.Params.Validate()
}
