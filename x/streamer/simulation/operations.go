package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Simulation operation weights constants
const (
	OpWeightSubmitProposal  = "op_weight_submit_proposal"
	OpWeightVoteProposal    = "op_weight_vote_proposal"
	OpWeightMsgUpdateParams = "op_weight_msg_update_params"

	DefaultWeightSubmitProposal      = 60
	DefaultWeightVoteProposal        = 40
	DefaultWeightMsgUpdateParams int = 10
)

// ProposalMsgs defines the module weighted proposals' contents
func ProposalMsgs() []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			OpWeightMsgUpdateParams,
			DefaultWeightMsgUpdateParams,
			SimulateMsgUpdateParams,
		),
	}
}

// SimulateMsgUpdateParams returns a random MsgUpdateParams
func SimulateMsgUpdateParams(r *rand.Rand, _ sdk.Context, _ []simtypes.Account) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	//params.BondDenom = simtypes.RandStringOfLength(r, 10)
	//params.HistoricalEntries = uint32(simtypes.RandIntBetween(r, 0, 1000))
	//params.MaxEntries = uint32(simtypes.RandIntBetween(r, 1, 1000))
	//params.MaxValidators = uint32(simtypes.RandIntBetween(r, 1, 1000))
	//params.UnbondingTime = time.Duration(simtypes.RandTimestamp(r).UnixNano())
	//params.MinCommissionRate = simtypes.RandomDecAmount(r, sdk.NewDec(1))

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simulation.WeightedOperations {
	var (
		weightSubmitProposal int
		weightVoteProposal   int
	)

	appParams.GetOrGenerate(cdc, OpWeightSubmitProposal, &weightSubmitProposal, nil,
		func(*rand.Rand) { weightSubmitProposal = DefaultWeightSubmitProposal })
	appParams.GetOrGenerate(cdc, OpWeightVoteProposal, &weightVoteProposal, nil,
		func(*rand.Rand) { weightVoteProposal = DefaultWeightVoteProposal })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightSubmitProposal,
			SimulateMsgSubmitProposal(protoCdc, ak, bk, ik, k),
		),
		simulation.NewWeightedOperation(
			weightVoteProposal,
			SimulateMsgVoteProposal(protoCdc, ak, bk, ik, k),
		),
	}
}

// SimulateMsgSubmitProposal generates random governance proposal content
func SimulateMsgSubmitProposal(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[r.Intn(len(accs))]

		// Randomly choose proposal type
		proposalTypes := []string{"create", "terminate", "replace", "update"}
		proposalType := proposalTypes[r.Intn(len(proposalTypes))]

		var content govtypes.Content
		var err error

		switch proposalType {
		case "create":
			content, err = generateCreateStreamProposal(r, ctx, ik)
		case "terminate":
			content, err = generateTerminateStreamProposal(r, ctx, k)
		case "replace":
			content, err = generateReplaceStreamProposal(r, ctx, k, ik)
		case "update":
			content, err = generateUpdateStreamProposal(r, ctx, k, ik)
		}

		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "submit_proposal", err.Error()), nil, nil
		}

		msg, err := govtypes.NewMsgSubmitProposal(content, sdk.NewCoins(), simAccount.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "submit_proposal", err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgCreateSponsoredStream generates a random sponsored stream creation
func SimulateMsgCreateSponsoredStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[r.Intn(len(accs))]

		// Get random gauges
		gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
		if len(gauges) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "no gauges"), nil, nil
		}

		// Generate random distribution records - for sponsored streams these don't matter
		// but we need valid ones for the API
		records := make([]types.DistrRecord, len(gauges))
		for i, gauge := range gauges {
			records[i] = types.DistrRecord{
				GaugeId: gauge.Id,
				Weight:  math.OneInt(),
			}
		}

		// Generate random coins similar to non-sponsored case
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "no spendable coins"), nil, nil
		}

		coin := spendable[r.Intn(len(spendable))]
		amt, err := dymsimtypes.RandIntBetween(r, math.OneInt(), coin.Amount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "failed to generate amount"), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(coin.Denom, amt))

		startTime := ctx.BlockTime().Add(time.Duration(r.Int63n(7*24*60*60)) * time.Second)
		numEpochs := uint64(r.Int63n(100) + 1)
		epochIdentifiers := []string{"day", "week", "month"}
		epochIdentifier := epochIdentifiers[r.Intn(len(epochIdentifiers))]

		streamID, err := k.CreateStream(ctx, coins, records, startTime, epochIdentifier, numEpochs, true)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgCreateStream{}, true, ""), nil, nil
	}
}

// SimulateMsgUpdateStream generates a random stream update
func SimulateMsgUpdateStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get all streams
		streams := k.GetStreams(ctx)
		if len(streams) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, "no streams"), nil, nil
		}

		// Pick a random stream
		stream := streams[r.Intn(len(streams))]

		// Get random gauges
		gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
		if len(gauges) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, "no gauges"), nil, nil
		}

		// Generate random distribution records
		records := make([]types.DistrRecord, len(gauges))
		totalWeight := math.ZeroInt()
		for i, gauge := range gauges {
			weight, err := dymsimtypes.RandIntBetween(r, math.OneInt(), math.NewIntFromUint64(^uint64(0)/uint64(len(gauges))))
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, "failed to generate weight"), nil, nil
			}
			records[i] = types.DistrRecord{
				GaugeId: gauge.Id,
				Weight:  weight,
			}
			totalWeight = totalWeight.Add(weight)
		}

		// Randomly choose between ReplaceDistrRecords and UpdateDistrRecords
		var err error
		if r.Int()%2 == 0 {
			err = k.ReplaceDistrRecords(ctx, stream.Id, records)
		} else {
			err = k.UpdateDistrRecords(ctx, stream.Id, records)
		}
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgUpdateStream{}, true, ""), nil, nil
	}
}

// SimulateMsgTerminateStream generates a random stream termination
func SimulateMsgTerminateStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get all streams
		streams := k.GetStreams(ctx)
		if len(streams) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTerminateStream, "no streams"), nil, nil
		}

		// Pick a random stream
		stream := streams[r.Intn(len(streams))]

		err := k.TerminateStream(ctx, stream.Id)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTerminateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgTerminateStream{}, true, ""), nil, nil
	}
}

// Helper functions for generating proposal content
func generateCreateStreamProposal(r *rand.Rand, ctx sdk.Context, ik dymsimtypes.IncentivesKeeper) (*types.CreateStreamProposal, error) {
	gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
	if len(gauges) == 0 {
		return nil, fmt.Errorf("no gauges available")
	}

	records := make([]types.DistrRecord, len(gauges))
	for i, gauge := range gauges {
		weight, err := dymsimtypes.RandIntBetween(r, math.OneInt(), math.NewIntFromUint64(^uint64(0)/uint64(len(gauges))))
		if err != nil {
			return nil, err
		}
		records[i] = types.DistrRecord{
			GaugeId: gauge.Id,
			Weight:  weight,
		}
	}

	coins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(simulation.RandIntBetween(r, 100, 10000)))))
	startTime := ctx.BlockTime().Add(time.Duration(r.Int63n(7*24*60*60)) * time.Second)
	epochIdentifiers := []string{"day", "week", "month"}

	return &types.CreateStreamProposal{
		Title:                simtypes.RandStringOfLength(r, 10),
		Description:          simtypes.RandStringOfLength(r, 100),
		DistributeToRecords:  records,
		Coins:                coins,
		StartTime:            startTime,
		DistrEpochIdentifier: epochIdentifiers[r.Intn(len(epochIdentifiers))],
		NumEpochsPaidOver:    uint64(r.Int63n(100) + 1),
		Sponsored:            r.Int()%2 == 0,
	}, nil
}

func generateTerminateStreamProposal(r *rand.Rand, ctx sdk.Context, k Keeper) (*types.TerminateStreamProposal, error) {
	streams := k.GetStreams(ctx)
	if len(streams) == 0 {
		return nil, fmt.Errorf("no streams available")
	}

	stream := streams[r.Intn(len(streams))]
	return &types.TerminateStreamProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    stream.Id,
	}, nil
}

func generateReplaceStreamProposal(r *rand.Rand, ctx sdk.Context, k Keeper, ik dymsimtypes.IncentivesKeeper) (*types.ReplaceStreamDistributionProposal, error) {
	streams := k.GetStreams(ctx)
	if len(streams) == 0 {
		return nil, fmt.Errorf("no streams available")
	}

	gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
	if len(gauges) == 0 {
		return nil, fmt.Errorf("no gauges available")
	}

	records := make([]types.DistrRecord, len(gauges))
	for i, gauge := range gauges {
		weight, err := dymsimtypes.RandIntBetween(r, math.OneInt(), math.NewIntFromUint64(^uint64(0)/uint64(len(gauges))))
		if err != nil {
			return nil, err
		}
		records[i] = types.DistrRecord{
			GaugeId: gauge.Id,
			Weight:  weight,
		}
	}

	stream := streams[r.Intn(len(streams))]
	return &types.ReplaceStreamDistributionProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    stream.Id,
		Records:     records,
	}, nil
}

func generateUpdateStreamProposal(r *rand.Rand, ctx sdk.Context, k Keeper, ik dymsimtypes.IncentivesKeeper) (*types.UpdateStreamDistributionProposal, error) {
	streams := k.GetStreams(ctx)
	if len(streams) == 0 {
		return nil, fmt.Errorf("no streams available")
	}

	gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
	if len(gauges) == 0 {
		return nil, fmt.Errorf("no gauges available")
	}

	records := make([]types.DistrRecord, len(gauges))
	for i, gauge := range gauges {
		weight, err := dymsimtypes.RandIntBetween(r, math.OneInt(), math.NewIntFromUint64(^uint64(0)/uint64(len(gauges))))
		if err != nil {
			return nil, err
		}
		records[i] = types.DistrRecord{
			GaugeId: gauge.Id,
			Weight:  weight,
		}
	}

	stream := streams[r.Intn(len(streams))]
	return &types.UpdateStreamDistributionProposal{
		Title:       simtypes.RandStringOfLength(r, 10),
		Description: simtypes.RandStringOfLength(r, 100),
		StreamId:    stream.Id,
		Records:     records,
	}, nil
}

// SimulateMsgVoteProposal simulates a vote on an active proposal
func SimulateMsgVoteProposal(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get random account
		simAccount := accs[r.Intn(len(accs))]

		// Get active proposals
		activeProposals := k.GetActiveProposals(ctx)
		if len(activeProposals) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "vote_proposal", "no active proposals"), nil, nil
		}

		// Pick random proposal
		proposal := activeProposals[r.Intn(len(activeProposals))]

		// Random vote option
		voteOptions := []govtypes.VoteOption{
			govtypes.OptionYes,
			govtypes.OptionNo,
			govtypes.OptionNoWithVeto,
			govtypes.OptionAbstain,
		}
		vote := voteOptions[r.Intn(len(voteOptions))]

		msg := govtypes.NewMsgVote(simAccount.Address, proposal.ProposalId, vote)

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}
