package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	keeper "github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
)

// Operation weights for simulating the module
const (
	OpWeightSubmitProposal = "op_weight_submit_proposal"
	OpWeightVoteProposal = "op_weight_vote_proposal"

	DefaultWeightSubmitProposal = 60
	DefaultWeightVoteProposal = 40
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightSubmitProposal int
		weightVoteProposal int
	)

	appParams.GetOrGenerate(cdc, OpWeightSubmitProposal, &weightSubmitProposal, nil,
		func(*rand.Rand) { weightSubmitProposal = DefaultWeightSubmitProposal })
	appParams.GetOrGenerate(cdc, OpWeightVoteProposal, &weightVoteProposal, nil,
		func(*rand.Rand) { weightVoteProposal = DefaultWeightVoteProposal })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightSubmitProposal,
			SimulateMsgSubmitProposal(k),
		),
		simulation.NewWeightedOperation(
			weightVoteProposal,
			SimulateMsgVoteProposal(k),
		),
	}
}

// SimulateMsgSubmitProposal simulates creating a new proposal
func SimulateMsgSubmitProposal(k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[r.Intn(len(accs))]
		proposal := generateRandomProposal(r)
		
		msg := govtypes.NewMsgSubmitProposal(
			proposal.Content,
			proposal.Deposit,
			simAccount.Address,
		)
		
		_, err := k.SubmitProposal(ctx, msg)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, proposal.Content.ProposalType(), "failed to submit proposal"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, proposal.Content.ProposalType()), nil, nil
	}
}

// SimulateMsgVoteProposal simulates voting on an existing proposal
func SimulateMsgVoteProposal(k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		proposals := k.GetProposals(ctx)
		if len(proposals) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "vote_proposal", "no proposals"), nil, nil
		}

		// Pick a random proposal
		proposal := proposals[r.Intn(len(proposals))]
		
		// Random vote option
		voteOptions := []govtypes.VoteOption{
			govtypes.OptionYes,
			govtypes.OptionNo,
			govtypes.OptionNoWithVeto,
			govtypes.OptionAbstain,
		}
		vote := voteOptions[r.Intn(len(voteOptions))]

		// Pick a random account to vote
		simAccount := accs[r.Intn(len(accs))]

		err := k.Vote(ctx, proposal.Id, simAccount.Address, vote)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "vote_proposal", "failed to vote"), nil, err
		}

		return simtypes.NewOperationMsg(&types.MsgVoteProposal{}, true, ""), nil, nil
	}
}

