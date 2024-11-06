package types

import "fmt"

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		SequencerList:    []Sequencer{},
		GenesisProposers: []GenesisProposer{},
		Params:           DefaultParams(),
		NoticeQueue:      []string{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in sequencer
	sequencerIndexMap := make(map[string]struct{})

	for _, elem := range gs.SequencerList {

		// TODO: should run validation on the sequencer objects

		seqKey := string(SequencerKey(elem.Address))
		if _, ok := sequencerIndexMap[seqKey]; ok {
			return fmt.Errorf("duplicated address for sequencer")
		}
		sequencerIndexMap[seqKey] = struct{}{}
	}

	if err := checkSecondIndex(gs.GenesisProposers, sequencerIndexMap); err != nil {
		return err
	}
	if err := checkSecondIndex(gs.GenesisSuccessors, sequencerIndexMap); err != nil {
		return err
	}

	for _, s := range gs.NoticeQueue {
		if _, ok := sequencerIndexMap[s]; !ok {
			return fmt.Errorf("notice queue contains non-existent sequencer")
		}
	}

	return gs.Params.ValidateBasic()
}

func checkSecondIndex(seqs []GenesisProposer, sequencerIndexMap map[string]struct{}) error {
	proposerIndexMap := make(map[string]struct{})
	for _, elem := range seqs {
		rollappId := elem.RollappId
		if _, ok := proposerIndexMap[rollappId]; ok {
			return fmt.Errorf("duplicated for %s", rollappId)
		}
		if _, ok := sequencerIndexMap[string(SequencerKey(elem.Address))]; !ok {
			return fmt.Errorf("%s does not have a sequencer", rollappId)
		}
		proposerIndexMap[rollappId] = struct{}{}
	}
	return nil
}
