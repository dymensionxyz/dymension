package app

import (
	"encoding/json"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/cosmos/cosmos-sdk/codec"
)

// The genesis state of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

// NewDefaultGenesisState generates the default state for the application.
func NewDefaultGenesisState(cdc codec.JSONCodec) GenesisState {
	defaultGenesisState := ModuleBasics.DefaultGenesis(cdc)

	if evmGenesisStateJson, found := defaultGenesisState[evmtypes.ModuleName]; found {
		// force disable Enable Create of x/evm
		var evmGenesisState evmtypes.GenesisState
		cdc.MustUnmarshalJSON(evmGenesisStateJson, &evmGenesisState)
		evmGenesisState.Params.EnableCreate = false
		defaultGenesisState[evmtypes.ModuleName] = cdc.MustMarshalJSON(&evmGenesisState)
	}

	return defaultGenesisState
}
