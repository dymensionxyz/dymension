package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/types/simulation"
	"time"
	"google.golang.org/protobuf/types/known/timestamppb"

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
	proposalTypes := []string{"CreateStream", "TerminateStream", "ReplaceStreamDistribution", "UpdateStreamDistribution"}
	proposalType := proposalTypes[r.Intn(len(proposalTypes))]

	title := fmt.Sprintf("Test %s Proposal", proposalType)
	description := fmt.Sprintf("This is a test %s proposal for simulation", proposalType)

	var content govtypes.Content
	switch proposalType {
	case "CreateStream":
		content = &types.CreateStreamProposal{
			Title:       title,
			Description: description,
			DistributeToRecords: []types.DistrRecord{
				{
					GaugeId: uint64(r.Int63n(100)),
					Weight:  sdk.NewInt(r.Int63n(100) + 1),
				},
			},
			Coins: sdk.NewCoins(sdk.NewCoin("udym", sdk.NewInt(r.Int63n(1000000)))),
			StartTime: timestamppb.New(time.Now().Add(
				time.Duration(r.Int63n(7*24)) * time.Hour)),
			DistrEpochIdentifier: "day",
			NumEpochsPaidOver:    uint64(r.Int63n(100) + 1),
			Sponsored:            r.Int()%2 == 0,
		}
	case "TerminateStream":
		content = &types.TerminateStreamProposal{
			Title:       title,
			Description: description,
			StreamId:    uint64(r.Int63n(100)),
		}
	case "ReplaceStreamDistribution":
		content = &types.ReplaceStreamDistributionProposal{
			Title:       title,
			Description: description,
			StreamId:    uint64(r.Int63n(100)),
			Records: []types.DistrRecord{
				{
					GaugeId: uint64(r.Int63n(100)),
					Weight:  sdk.NewInt(r.Int63n(100) + 1),
				},
			},
		}
	case "UpdateStreamDistribution":
		content = &types.UpdateStreamDistributionProposal{
			Title:       title,
			Description: description,
			StreamId:    uint64(r.Int63n(100)),
			Records: []types.DistrRecord{
				{
					GaugeId: uint64(r.Int63n(100)),
					Weight:  sdk.NewInt(r.Int63n(100) + 1),
				},
			},
		}
	}

	return types.Proposal{
		Content:      content,
		Id:           uint64(r.Int63n(1000000)),
		Status:       govtypes.StatusDepositPeriod,
		VotingPeriod: time.Duration(r.Intn(13*24)+24) * time.Hour,
		Deposit:      sdk.NewCoin("udym", sdk.NewInt(r.Int63n(1000000000))),
	}
}
