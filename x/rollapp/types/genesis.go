package types

import (
	"errors"
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
		AppList:                            []App{},
		Params:                             DefaultParams(),
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
			return errors.New("duplicated index for rollapp")
		}
		rollappIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in stateInfo
	stateInfoIndexMap := make(map[string]struct{})

	for _, elem := range gs.StateInfoList {
		index := string(StateInfoKey(elem.StateInfoIndex))
		if _, ok := stateInfoIndexMap[index]; ok {
			return errors.New("duplicated index for stateInfo")
		}
		stateInfoIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in latestStateInfoIndex
	latestStateInfoIndexIndexMap := make(map[string]struct{})

	for _, elem := range gs.LatestStateInfoIndexList {
		index := string(LatestStateInfoIndexKey(elem.RollappId))
		if _, ok := latestStateInfoIndexIndexMap[index]; ok {
			return errors.New("duplicated index for latestStateInfoIndex")
		}
		latestStateInfoIndexIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in latestFinalizedStateIndex
	latestFinalizedStateIndexIndexMap := make(map[string]struct{})

	for _, elem := range gs.LatestFinalizedStateIndexList {
		index := string(LatestFinalizedStateIndexKey(elem.RollappId))
		if _, ok := latestFinalizedStateIndexIndexMap[index]; ok {
			return errors.New("duplicated index for latestFinalizedStateIndex")
		}
		latestFinalizedStateIndexIndexMap[index] = struct{}{}
	}
	// Check for duplicated index in blockHeightToFinalizationQueue
	blockHeightToFinalizationQueueIndexMap := make(map[string]struct{})

	for _, elem := range gs.BlockHeightToFinalizationQueueList {
		index := string(BlockHeightToFinalizationQueueKey(elem.CreationHeight))
		if _, ok := blockHeightToFinalizationQueueIndexMap[index]; ok {
			return errors.New("duplicated index for blockHeightToFinalizationQueue")
		}
		blockHeightToFinalizationQueueIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in app
	appIndexMap := make(map[string]struct{})

	for _, elem := range gs.AppList {
		index := string(AppKey(elem))
		if _, ok := appIndexMap[index]; ok {
			return errors.New("duplicated index for app")
		}
		appIndexMap[index] = struct{}{}
	}

	return gs.Params.Validate()
}
