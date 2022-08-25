package types

import (
	"fmt"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		SequencerList:           []Sequencer{},
		SequencersByRollappList: []SequencersByRollapp{},
		SchedulerList: []Scheduler{},
// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in sequencer
	sequencerIndexMap := make(map[string]struct{})

	for _, elem := range gs.SequencerList {
		index := string(SequencerKey(elem.SequencerAddress))
		if _, ok := sequencerIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for sequencer")
		}
		sequencerIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in sequencersByRollapp
	sequencersByRollappIndexMap := make(map[string]struct{})

	for _, elem := range gs.SequencersByRollappList {
		index := string(SequencersByRollappKey(elem.RollappId))
		if _, ok := sequencersByRollappIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for sequencersByRollapp")
		}
		sequencersByRollappIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in scheduler
schedulerIndexMap := make(map[string]struct{})

for _, elem := range gs.SchedulerList {
	index := string(SchedulerKey(elem.SequencerAddress))
	if _, ok := schedulerIndexMap[index]; ok {
		return fmt.Errorf("duplicated index for scheduler")
	}
	schedulerIndexMap[index] = struct{}{}
}
// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
