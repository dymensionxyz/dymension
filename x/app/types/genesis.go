package types

import (
	"errors"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		AppList: []App{},
		Params:  DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in app
	appIndexMap := make(map[string]struct{})

	for _, elem := range gs.AppList {
		index := string(AppKey(elem.Name, elem.RollappId))
		if _, ok := appIndexMap[index]; ok {
			return errors.New("duplicated index for app")
		}
		appIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
