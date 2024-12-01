package simulation

import (
	"fmt"
	"math/rand"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgTestBondingCurve int = 100
	OpWeightMsgTestBondingCurve          = "op_weight_msg_test_bonding_curve"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {

	protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	var weightMsgTestBondingCurve int
	appParams.GetOrGenerate(cdc, OpWeightMsgTestBondingCurve, &weightMsgTestBondingCurve, nil,
		func(_ *rand.Rand) {
			weightMsgTestBondingCurve = DefaultWeightMsgTestBondingCurve
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgTestBondingCurve,
			SimulateTestBondingCurve(k, protoCdc),
		),
	}
}

// SimulateTestBondingCurve tests the bonding curve calculations without actual trading
func SimulateTestBondingCurve(k keeper.Keeper, cdc *codec.ProtoCodec) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := k.GetAllPlans(ctx, true)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "TestBondingCurve", "no plans available"), nil, nil
		}

		// Randomly select a plan
		plan := plans[r.Intn(len(plans))]
		curve := plan.BondingCurve

		// Test different token amounts
		testAmounts := []math.Int{
			math.NewInt(1).MulRaw(1e18),      // 1 token
			math.NewInt(100).MulRaw(1e18),    // 100 tokens
			math.NewInt(1000).MulRaw(1e18),   // 1000 tokens
			math.NewInt(10000).MulRaw(1e18),  // 10000 tokens
			math.NewInt(100000).MulRaw(1e18), // 100000 tokens
		}

		var results []string
		for _, amount := range testAmounts {
			// Calculate cost for buying tokens
			cost := curve.Cost(plan.SoldAmt, amount)

			// Calculate tokens for exact DYM spend
			tokens, err := curve.TokensForExactDYM(plan.SoldAmt, cost)
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, "TestBondingCurve", "TokensForExactDYM failed"), nil, err
			}

			// Calculate spot price at this amount
			spotPrice := curve.SpotPrice(amount)

			results = append(results, fmt.Sprintf(
				"Amount: %s tokens, Cost: %s DYM, TokensForExactDYM: %s, SpotPrice: %s",
				amount, cost, tokens, spotPrice,
			))
		}

		return simtypes.NewOperationMsg(&types.MsgBuy{}, true, fmt.Sprintf("TestBondingCurve Results:\n%s", results), cdc), nil, nil
	}
}
