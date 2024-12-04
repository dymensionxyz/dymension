package simulation

import (
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

const (
	WeightCreateStreamProposal              = 100
	WeightTerminateStreamProposal           = 100
	WeightReplaceStreamDistributionProposal = 100
	WeightUpdateStreamDistributionProposal  = 100
	WeightFundModule                        = 100
)

type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	types.BankKeeper
}

type EpochKeeper interface {
	types.EpochKeeper
}

type AccountKeeper interface {
	types.AccountKeeper
}

type IncentivesKeeper interface {
	GetGauges(ctx sdk.Context) []incentivestypes.Gauge
	types.IncentivesKeeper
}

type SponsorshipKeeper interface {
	types.SponsorshipKeeper
}

type Keepers struct {
	Bank       BankKeeper
	Epoch      EpochKeeper
	Acc        AccountKeeper
	Incentives IncentivesKeeper
	Endorse    SponsorshipKeeper
}

type OpFactory struct {
	*keeper.Keeper
	k Keepers
	module.SimulationState
}

func NewOpFactory(k *keeper.Keeper, ks Keepers, simState module.SimulationState) OpFactory {
	return OpFactory{
		Keeper:          k,
		k:               ks,
		SimulationState: simState,
	}
}

func (f OpFactory) Messages() []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{
		//simulation.NewWeightedOperation(
		//	WeightFundModule,
		//	f.FundModule,
		//),
	}
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

//func (f OpFactory) FundModule(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accounts []simtypes.Account, id string) (OperationMsg simtypes.OperationMsg, futureOps []simtypes.FutureOperation, err error) {
//
//}

func (f *OpFactory) CreateStreamProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	var epoch string
	{
		epochs := f.k.Epoch.AllEpochInfos(ctx)
		if 0 < len(epochs) {
			epoch = dymsimtypes.RandChoice(r, epochs).Identifier
		}
	}

	var coins sdk.Coins
	{
		// NOTE: I'm pretty sure this test is a no op because nothing is funding this module
		bal := f.k.Bank.GetAllBalances(ctx, f.k.Acc.GetModuleAddress(types.ModuleName))
		coins = simtypes.RandSubsetCoins(r, bal)
	}
	if coins.Empty() {
		return nil
	}

	records := f.GetDistr(r, ctx)

	return &types.CreateStreamProposal{
		Title:                simtypes.RandStringOfLength(r, 10),
		Description:          simtypes.RandStringOfLength(r, 100),
		DistributeToRecords:  records,
		Coins:                coins,
		StartTime:            dymsimtypes.RandFutureTime(r, ctx, time.Minute),
		DistrEpochIdentifier: epoch,
		NumEpochsPaidOver:    uint64(simtypes.RandIntBetween(r, 0, 100)),
		Sponsored:            r.Int()%2 == 0,
	}
}

func (f *OpFactory) TerminateStreamProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	s := f.GetStream(r, ctx)
	if s == nil {
		return nil
	}

	return &types.TerminateStreamProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    *s,
	}
}

func (f *OpFactory) ReplaceStreamDistributionProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	s := f.GetStream(r, ctx)
	if s == nil {
		return nil
	}
	distr := f.GetDistr(r, ctx)
	return &types.ReplaceStreamDistributionProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    *s,
		Records:     distr,
	}
}

func (f *OpFactory) UpdateStreamDistributionProposal(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) simtypes.Content {
	s := f.GetStream(r, ctx)
	if s == nil {
		return nil
	}
	distr := f.GetDistr(r, ctx)

	return &types.UpdateStreamDistributionProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    *s,
		Records:     distr,
	}
}

func (f *OpFactory) GetStream(r *rand.Rand, ctx sdk.Context) *uint64 {
	ss := f.GetStreams(ctx)
	if len(ss) == 0 {
		return nil
	}
	x := dymsimtypes.RandChoice(r, ss).Id
	return &x
}

func (f *OpFactory) GetDistr(r *rand.Rand, ctx sdk.Context) []types.DistrRecord {
	gauges := dymsimtypes.RandomGaugeSubset(ctx, r, f.k.Incentives)
	records := make([]types.DistrRecord, len(gauges))
	for _, gauge := range gauges {
		records = append(records, types.DistrRecord{
			GaugeId: gauge.Id,
			Weight:  sdk.NewInt(int64(simtypes.RandIntBetween(r, 1, 100))),
		})
	}
	return records
}
