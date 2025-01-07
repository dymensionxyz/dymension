package types

import (
	"errors"
	fmt "fmt"
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
	blockHeightToFinalizationQueueIndexMap := make(map[uint64]struct{})

	for _, elem := range gs.BlockHeightToFinalizationQueueList {
		index := elem.CreationHeight
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

	// Check for duplicated index in obsolete DRS versions
	obsoleteDRSVersionIndexMap := make(map[uint32]struct{})

	// Check for duplicated index in livenessEvents
	livenessEventsIndexMap := make(map[string]struct{})
	for _, elem := range gs.LivenessEvents {
		index := elem.RollappId
		if _, ok := livenessEventsIndexMap[index]; ok {
			return errors.New("duplicated index for LivenessEvents")
		}
		livenessEventsIndexMap[index] = struct{}{}
	}

	// Check for duplicated index in registerDenoms
	registeredDenomsIndexMap := make(map[string]struct{})
	for _, entry := range gs.RegisteredDenoms {
		if entry.RollappId == "" {
			return errors.New("invalid RegisteredDenoms: RollappId cannot be empty")
		}

		if _, ok := registeredDenomsIndexMap[entry.RollappId]; ok {
			return fmt.Errorf("duplicate RegisteredDenoms entry for RollappId: %s", entry.RollappId)
		}
		registeredDenomsIndexMap[entry.RollappId] = struct{}{}

		if len(entry.Denoms) == 0 {
			return fmt.Errorf("invalid RegisteredDenoms for RollappId %s: Denoms list cannot be empty", entry.RollappId)
		}

		denomSet := make(map[string]struct{})
		for _, denom := range entry.Denoms {
			if denom == "" {
				return fmt.Errorf("invalid RegisteredDenoms for RollappId %s: Denom cannot be empty", entry.RollappId)
			}
			if _, exists := denomSet[denom]; exists {
				return fmt.Errorf("duplicate Denom '%s' found in RollappId: %s", denom, entry.RollappId)
			}
			denomSet[denom] = struct{}{}
		}
	}

	// Check for duplicated index in SequencerHeightPairs
	sequencerHeightPairsIndexMap := make(map[string]struct{})

	for _, pair := range gs.SequencerHeightPairs {
		if pair.Sequencer == "" {
			return errors.New("invalid SequencerHeightPair: Sequencer cannot be empty")
		}

		if pair.Height == 0 {
			return errors.New("invalid SequencerHeightPair: Height must be greater than 0")
		}

		index := fmt.Sprintf("%s-%d", pair.Sequencer, pair.Height)
		if _, exists := sequencerHeightPairsIndexMap[index]; exists {
			return fmt.Errorf("duplicated SequencerHeightPair: Sequencer '%s' with Height '%d'", pair.Sequencer, pair.Height)
		}
		sequencerHeightPairsIndexMap[index] = struct{}{}
	}

	for _, elem := range gs.ObsoleteDrsVersions {
		if _, ok := obsoleteDRSVersionIndexMap[elem]; ok {
			return errors.New("duplicated index for ObsoleteDrsVersions")
		}
		obsoleteDRSVersionIndexMap[elem] = struct{}{}
	}

	return gs.Params.Validate()
}
