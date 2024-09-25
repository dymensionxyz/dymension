package types

import fmt "fmt"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	rollapps := make(map[string]bool)
	ids := make(map[uint64]bool)

	for _, plan := range gs.Plans {
		if err := plan.ValidateBasic(); err != nil {
			return err
		}

		if _, found := rollapps[plan.RollappId]; found {
			return fmt.Errorf("duplicate rollapp ID %s", plan.RollappId)
		}
		rollapps[plan.RollappId] = true

		if _, found := ids[plan.Id]; found {
			return fmt.Errorf("duplicate plan ID %d", plan.Id)
		}
		ids[plan.Id] = true
	}

	return gs.Params.Validate()
}
