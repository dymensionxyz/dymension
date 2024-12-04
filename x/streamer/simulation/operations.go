package simulation

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)


const (
	WeightCreateStreamProposal              = 100
	WeightTerminateStreamProposal           = 100
	WeightReplaceStreamDistributionProposal = 100
	WeightUpdateStreamDistributionProposal  = 100
)

type OpFactory struct {
	*keeper.Keeper
	module.SimulationState
}

func

func (f OpFactory) Proposals() []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{
		simulation.NewWeightedProposalContent(
			"op_create_stream_proposal",
			WeightCreateStreamProposal,
			f.CreateStreamProposal,
		),
		simulation.NewWeightedProposalContent(
			"op_terminate_stream_proposal",
			WeightTerminateStreamProposal,
			f.TerminateStreamProposal,
		),
		simulation.NewWeightedProposalContent(
			"op_replace_stream_distribution_proposal",
			WeightReplaceStreamDistributionProposal,
			f.ReplaceStreamDistributionProposal,
		),
		simulation.NewWeightedProposalContent(
			"op_update_stream_distribution_proposal",
			WeightUpdateStreamDistributionProposal,
			f.UpdateStreamDistributionProposal,
		),
	}
}

func (f *OpFactory) CreateStreamProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	f.
	// Generate random distribution records
	numRecords := simtypes.RandIntBetween(r, 1, 5)
	records := make([]types.DistrRecord, numRecords)
	for i := 0; i < numRecords; i++ {
		records[i] = types.DistrRecord{
			GaugeId: uint64(simtypes.RandIntBetween(r, 1, 100)),
			Weight:  sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 100))),
		}
	}

	// Generate random coins
	amount := sdk.NewInt(int64(simtypes.RandIntBetween(r, 100, 10000)))

	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, amount))

	// Generate random start time between now and 1 week in the future
	startTime := ctx.BlockTime().Add(time.Duration(r.Int63n(7*24*60*60)) * time.Second) // TODO: does it do anything

	// Random epoch identifier
	epochIdentifiers := []string{"day", "week", "month"}
	epochIdentifier := epochIdentifiers[r.Intn(len(epochIdentifiers))]

	return &types.CreateStreamProposal{
		Title:                simtypes.RandStringOfLength(r, 10),
		Description:          simtypes.RandStringOfLength(r, 100),
		DistributeToRecords:  records,
		Coins:                coins,
		StartTime:            startTime,
		DistrEpochIdentifier: epochIdentifier,
		NumEpochsPaidOver:    uint64(simtypes.RandIntBetween(r, 1, 100)),
		Sponsored:            r.Int()%2 == 0,
	}
}

func (f *OpFactory) TerminateStreamProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	// Generate random stream ID between 1 and 1000
	streamId := uint64(simtypes.RandIntBetween(r, 1, 1000))

	return &types.TerminateStreamProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    streamId,
	}
}

func (f *OpFactory) ReplaceStreamDistributionProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	// Generate random stream ID
	streamId := uint64(simtypes.RandIntBetween(r, 1, 1000))

	// Generate random distribution records
	numRecords := simtypes.RandIntBetween(r, 1, 5)
	records := make([]types.DistrRecord, numRecords)
	for i := 0; i < numRecords; i++ {
		records[i] = types.DistrRecord{
			GaugeId: uint64(simtypes.RandIntBetween(r, 1, 100)),
			Weight:  sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 100))),
		}
	}

	return &types.ReplaceStreamDistributionProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    streamId,
		Records:     records,
	}
}

func (f *OpFactory) UpdateStreamDistributionProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	// Generate random stream ID
	streamId := uint64(simtypes.RandIntBetween(r, 1, 1000))

	// Generate random distribution records
	numRecords := simtypes.RandIntBetween(r, 1, 5)
	records := make([]types.DistrRecord, numRecords)
	for i := 0; i < numRecords; i++ {
		records[i] = types.DistrRecord{
			GaugeId: uint64(simtypes.RandIntBetween(r, 1, 100)),
			Weight:  sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 100))),
		}
	}

	return &types.UpdateStreamDistributionProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    streamId,
		Records:     records,
	}
}
