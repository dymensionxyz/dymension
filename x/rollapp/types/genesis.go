package types

import (
	"fmt"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		RollappList:    []Rollapp{},
		StateInfoList:  []StateInfo{},
		StateIndexList: []StateIndex{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in rollapp
	rollappIndexMap := make(map[string]struct{})

	for _, elem := range gs.RollappList {
		index := string(RollappKey(elem.RollappId))
		if _, ok := rollappIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for rollapp")
		}
		rollappIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in stateInfo
	stateInfoIndexMap := make(map[string]struct{})

	for _, elem := range gs.StateInfoList {
		index := string(StateInfoKey(elem.RollappId, elem.StateIndex))
		if _, ok := stateInfoIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for stateInfo")
		}
		stateInfoIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in stateIndex
	stateIndexIndexMap := make(map[string]struct{})

	for _, elem := range gs.StateIndexList {
		index := string(StateIndexKey(elem.RollappId))
		if _, ok := stateIndexIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for stateIndex")
		}
		stateIndexIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
