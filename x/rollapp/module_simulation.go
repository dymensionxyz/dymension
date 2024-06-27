package rollapp

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	rollappsimulation "github.com/dymensionxyz/dymension/v3/x/rollapp/simulation"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = params.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	// nolint: gosec
	opWeightMsgCreateRollapp = "op_weight_msg_create_rollapp"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateRollapp int = 100

	// nolint: gosec
	opWeightMsgUpdateState = "op_weight_msg_update_state"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateState int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	rollappGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&rollappGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreateRollapp int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateRollapp, &weightMsgCreateRollapp, nil,
		func(_ *rand.Rand) {
			weightMsgCreateRollapp = defaultWeightMsgCreateRollapp
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateRollapp,
		rollappsimulation.SimulateMsgCreateRollapp(am.accountKeeper, am.bankKeeper),
	))

	var weightMsgUpdateState int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateState, &weightMsgUpdateState, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateState = defaultWeightMsgUpdateState
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateState,
		rollappsimulation.SimulateMsgUpdateState(am.accountKeeper, am.bankKeeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
