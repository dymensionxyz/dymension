package simulation

import (
	"math/rand"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// Simulation operation weights constants.
const (
	DefaultWeightMsgVote       int = 100
	DefaultWeightMsgRevokeVote int = 100
	OpWeightMsgVote                = "op_weight_msg_vote"        //nolint:gosec
	OpWeightMsgRevokeVote          = "op_weight_msg_revoke_vote" //nolint:gosec
)

// WeightedOperations returns all the operations from the module with their respective weights.
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	sk dymsimtypes.StakingKeeper,
	s keeper.Keeper,
) simulation.WeightedOperations {
	var weightMsgVote int

	appParams.GetOrGenerate(
		cdc, OpWeightMsgVote, &weightMsgVote, nil,
		func(*rand.Rand) { weightMsgVote = DefaultWeightMsgVote },
	)

	protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgVote,
			SimulateMsgVote(protoCdc, ak, bk, ik, sk, s),
		),
	}
}

// getAllocationWeight returns a random allocation weight in range [minAllocationWeight; MaxAllocationWeight].
func getAllocationWeight(r *rand.Rand, minAllocationWeight math.Int) math.Int {
	w, _ := dymsimtypes.RandIntBetween(r, minAllocationWeight, types.MaxAllocationWeight.AddRaw(1))
	return w
}

// SimulateMsgVote generates and executes a MsgVote with random parameters
func SimulateMsgVote(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	sk dymsimtypes.StakingKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		params, _ := k.GetParams(ctx)

		delegation := dymsimtypes.RandomDelegation(ctx, r, sk)
		if delegation == nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgVote, "No delegation available"), nil, nil
		}

		b, err := k.GetValidatorBreakdown(ctx, delegation.GetDelegatorAddr())
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgVote, "Failed to get validator breakdown"), nil, err
		}

		if b.TotalPower.LT(params.MinVotingPower) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgVote, "Address does not have enough staking power to vote"), nil, nil
		}

		// Get a random subset of gauges
		selectedGauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
		if len(selectedGauges) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgVote, "No gauges available"), nil, nil
		}

		// Generate random weights for the selected gauges.
		// The sum of the weights should be less than or equal to 100 DYM (100%).
		totalWeight := math.ZeroInt()
		var gaugeWeights []types.GaugeWeight
		for _, gauge := range selectedGauges {
			weight := getAllocationWeight(r, params.MinAllocationWeight)
			if totalWeight.Add(weight).GT(types.MaxAllocationWeight) {
				weight = types.MaxAllocationWeight.Sub(totalWeight)
			}

			if weight.LT(params.MinAllocationWeight) {
				// We don't have any more weight to distribute.
				// The remaining weight is abstained.
				break
			}

			gaugeWeights = append(gaugeWeights, types.GaugeWeight{
				GaugeId: gauge.Id,
				Weight:  weight,
			})

			totalWeight = totalWeight.Add(weight)
		}

		msg := &types.MsgVote{
			Voter:   delegation.GetDelegatorAddr().String(),
			Weights: gaugeWeights,
		}

		// Need to retrieve the simulation account associated with delegation to retrieve PrivKey
		var simAccount simtypes.Account

		for _, simAcc := range accs {
			if simAcc.Address.Equals(delegation.GetDelegatorAddr()) {
				simAccount = simAcc
				break
			}
		}
		// If simaccount.PrivKey == nil, delegation address does not exist in accs. However, since smart contracts and module accounts can stake, we can ignore the error.
		if simAccount.PrivKey == nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "Voter account private key is nil"), nil, nil
		}

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:             cdc,
			Msg:             msg,
			MsgType:         msg.Type(),
			CoinsSpentInMsg: nil,
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
