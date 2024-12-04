package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Simulation parameter constants
const (
	MaxIterationsPerBlock = "max_iterations_per_block"
	InitialProposalCount = "initial_proposal_count"
)

// Default simulation parameters
const (
	DefaultMinInitialProposalCount = 0
	DefaultMaxInitialProposalCount = 5
)

// GenMaxIterationsPerBlock randomized MaxIterationsPerBlock
func GenMaxIterationsPerBlock(r *rand.Rand) uint64 {
	return uint64(simulation.RandIntBetween(r, 100, 10000))
}

// RandomizedGenState generates a random GenesisState for streamer module
func RandomizedGenState(simState *module.SimulationState) {
	var maxIterationsPerBlock uint64

	simState.AppParams.GetOrGenerate(
		simState.Cdc, MaxIterationsPerBlock, &maxIterationsPerBlock, simState.Rand,
		func(r *rand.Rand) { maxIterationsPerBlock = GenMaxIterationsPerBlock(r) },
	)

	var initialProposalCount int
	simState.AppParams.GetOrGenerate(
		simState.Cdc, InitialProposalCount, &initialProposalCount, simState.Rand,
		func(r *rand.Rand) {
			initialProposalCount = r.Intn(DefaultMaxInitialProposalCount-DefaultMinInitialProposalCount) + DefaultMinInitialProposalCount
		},
	)

	// Generate random proposals
	proposals := make([]types.Proposal, initialProposalCount)
	for i := 0; i < initialProposalCount; i++ {
		proposals[i] = generateRandomProposal(simState.Rand)
	}

	streamerGenesis := types.GenesisState{
		Params: types.Params{
			MaxIterationsPerBlock: maxIterationsPerBlock,
		},
		Streams:       []types.Stream{},
		LastStreamId:  0,
		EpochPointers: []types.EpochPointer{},
		Proposals:     proposals,
	}

	bz, err := json.MarshalIndent(&streamerGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated streamer parameters:\n%s\n", bz)

	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&streamerGenesis)
}

// generateRandomProposal creates a random proposal for simulation
func generateRandomProposal(r *rand.Rand) types.Proposal {
	// Generate random proposal parameters
	proposalID := uint64(r.Int63n(1000000))
	status := govtypes.StatusDepositPeriod
	
	// Random voting period between 1 day and 2 weeks
	votingPeriod := time.Duration(r.Intn(13*24)+24) * time.Hour
	
	// Random deposit
	depositAmount := sdk.NewInt(r.Int63n(1000000000))
	deposit := sdk.NewCoin("udym", depositAmount)

	return types.Proposal{
		Id:            proposalID,
		Status:        status,
		VotingPeriod: votingPeriod,
		Deposit:      deposit,
	}
}
