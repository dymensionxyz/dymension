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
		Params:              DefaultParams(),
		Denommetadatas:      []DenomMetadata{},
		LastDenommetadataId: 0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}

	if len(gs.Denommetadatas) != int(gs.LastDenommetadataId) {
		return fmt.Errorf("denommetadata length does not match last denommetadata id")
	}

	//validate the denommetadatas are sorted and last denommetadata id is correct
	for i, denommetadata := range gs.Denommetadatas {
		if denommetadata.Id != uint64(i+1) {
			return fmt.Errorf("denommetadata in idx %d have wrong denomID (%d)", i, denommetadata.Id)
		}
	}

	return nil
}
