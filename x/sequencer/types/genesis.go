package types

import fmt "fmt"

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		SequencerList:    []Sequencer{},
		GenesisProposers: []GenesisProposer{},
		Params:           DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in sequencer
	sequencerIndexMap := make(map[string]struct{})

	for _, elem := range gs.SequencerList {

		// TODO: should run validation on the sequencer objects

		seqKey := string(SequencerKey(elem.SequencerAddress))
		if _, ok := sequencerIndexMap[seqKey]; ok {
			return fmt.Errorf("duplicated index for sequencer")
		}
		sequencerIndexMap[seqKey] = struct{}{}
	}

	// Check for duplicated index in proposer
	proposerIndexMap := make(map[string]struct{})
	for _, elem := range gs.GenesisProposers {
		rollappId := string(elem.RollappId)
		if _, ok := proposerIndexMap[rollappId]; ok {
			return fmt.Errorf("duplicated proposer for %s", rollappId)
		}
		if _, ok := sequencerIndexMap[string(SequencerKey(elem.Address))]; !ok {
			return fmt.Errorf("proposer %s does not have a sequencer", rollappId)
		}
		proposerIndexMap[rollappId] = struct{}{}
	}

	return gs.Params.Validate()
}
