package types

import (
	"fmt"
)

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		SequencerList:           []Sequencer{},
		SequencersByRollappList: []SequencersByRollapp{},
		Params:                  DefaultParams(),
	}
}

//FIXME: should run validation on the sequencer objects

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

	//FIXME: validate single PROPOSER per rollapp

	return gs.Params.Validate()
}
