package types

import (
	"fmt"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		RollappList:                        []Rollapp{},
		StateInfoList:                      []StateInfo{},
		LatestStateInfoIndexList:           []StateInfoIndex{},
		LatestFinalizedStateIndexList:      []StateInfoIndex{},
		BlockHeightToFinalizationQueueList: []BlockHeightToFinalizationQueue{},
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
		index := string(StateInfoKey(elem.StateInfoIndex))
		if _, ok := stateInfoIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for stateInfo")
		}
		stateInfoIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in latestStateInfoIndex
	latestStateInfoIndexIndexMap := make(map[string]struct{})

	for _, elem := range gs.LatestStateInfoIndexList {
		index := string(LatestStateInfoIndexKey(elem.RollappId))
		if _, ok := latestStateInfoIndexIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for latestStateInfoIndex")
		}
		latestStateInfoIndexIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in latestFinalizedStateIndex
	latestFinalizedStateIndexIndexMap := make(map[string]struct{})

	for _, elem := range gs.LatestFinalizedStateIndexList {
		index := string(LatestFinalizedStateIndexKey(elem.RollappId))
		if _, ok := latestFinalizedStateIndexIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for latestFinalizedStateIndex")
		}
		latestFinalizedStateIndexIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in blockHeightToFinalizationQueue
	blockHeightToFinalizationQueueIndexMap := make(map[string]struct{})

	for _, elem := range gs.BlockHeightToFinalizationQueueList {
		index := string(BlockHeightToFinalizationQueueKey(elem.FinalizationHeight))
		if _, ok := blockHeightToFinalizationQueueIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for blockHeightToFinalizationQueue")
		}
		blockHeightToFinalizationQueueIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
