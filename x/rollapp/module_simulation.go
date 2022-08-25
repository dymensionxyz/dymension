package rollapp

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/testutil/sample"
	rollappsimulation "github.com/dymensionxyz/dymension/x/rollapp/simulation"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = rollappsimulation.FindAccount
	_ = simappparams.StakePerAccount
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

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	rollappParams := types.DefaultParams()
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyDisputePeriodInBlocks), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(rollappParams.DisputePeriodInBlocks))
		}),
	}
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
		rollappsimulation.SimulateMsgCreateRollapp(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateState int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateState, &weightMsgUpdateState, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateState = defaultWeightMsgUpdateState
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateState,
		rollappsimulation.SimulateMsgUpdateState(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
