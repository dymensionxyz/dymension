package types

import (
	"encoding/json"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
)

// DefaultIndex is the default incentive module's global index.
const DefaultIndex uint64 = 1

// DefaultGenesis returns the incentive module's default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Gauges: []Gauge{},
		LockableDurations: []time.Duration{
			time.Second,
			time.Hour,
			time.Hour * 3,
			time.Hour * 7,
		},
	}
}

// GetGenesisStateFromAppState returns an incentive module GenesisState given a raw application genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// Validate performs basic genesis state validation, returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}

	// TODO: validate gauges

	return nil
}
