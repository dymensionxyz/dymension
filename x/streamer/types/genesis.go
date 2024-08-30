package types

import (
	"fmt"
)

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		Streams:       []Stream{},
		LastStreamId:  0,
		EpochPointers: []EpochPointer{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}

	if len(gs.Streams) != int(gs.LastStreamId) {
		return fmt.Errorf("streams length does not match last stream id")
	}

	// validate the streams are sorted and last stream id is correct
	for i, stream := range gs.Streams {
		if stream.Id != uint64(i+1) {
			return fmt.Errorf("stream in idx %d have wrong streamID (%d)", i, stream.Id)
		}
	}

	return nil
}
