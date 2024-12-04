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
	return &types.CreateStreamProposal{
		Title:                "",
		Description:          "",
		DistributeToRecords:  nil,
		Coins:                nil,
		StartTime:            time.Time{},
		DistrEpochIdentifier: "",
		NumEpochsPaidOver:    0,
		Sponsored:            false,
	}
}

func (f *OpFactory) TerminateStreamProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	return &types.TerminateStreamProposal{
		Title:       "",
		Description: "",
		StreamId:    0,
	}
}

func (f *OpFactory) ReplaceStreamDistributionProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	return &types.ReplaceStreamDistributionProposal{
		Title:       "",
		Description: "",
		StreamId:    0,
		Records:     nil,
	}
}

func (f *OpFactory) UpdateStreamDistributionProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	return &types.UpdateStreamDistributionProposal{
		Title:       "",
		Description: "",
		StreamId:    0,
		Records:     nil,
	}
}
