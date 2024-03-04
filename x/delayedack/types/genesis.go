package types

import (
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:         DefaultParams(),
		RollappPackets: []commontypes.RollappPacket{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
