package simulation

import (
	"fmt"
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
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// Simulation operation weights constants
const (
	DefaultWeightMsgBuy           int = 100
	DefaultWeightMsgBuyExactSpend int = 100
	DefaultWeightMsgSell          int = 100

	OpWeightMsgBuy           = "op_weight_msg_buy"
	OpWeightMsgBuyExactSpend = "op_weight_msg_buy_exact_spend"
	OpWeightMsgSell          = "op_weight_msg_sell"
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgBuy           int
		weightMsgBuyExactSpend int
		weightMsgSell          int
	)

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	protoCdc := codec.NewProtoCodec(interfaceRegistry)

	appParams.GetOrGenerate(
		cdc, OpWeightMsgBuy, &weightMsgBuy, nil,
		func(*rand.Rand) { weightMsgBuy = DefaultWeightMsgBuy },
	)

	appParams.GetOrGenerate(
		cdc, OpWeightMsgBuyExactSpend, &weightMsgBuyExactSpend, nil,
		func(*rand.Rand) { weightMsgBuyExactSpend = DefaultWeightMsgBuyExactSpend },
	)

	appParams.GetOrGenerate(
		cdc, OpWeightMsgSell, &weightMsgSell, nil,
		func(*rand.Rand) { weightMsgSell = DefaultWeightMsgSell },
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgBuy,
			SimulateMsgBuy(protoCdc, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgBuyExactSpend,
			SimulateMsgBuyExactSpend(protoCdc, ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgSell,
			SimulateMsgSell(protoCdc, ak, bk, k),
		),
	}
}

// SimulateMsgBuy generates a MsgBuy with random values
func SimulateMsgBuy(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)

		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBuy, "no spendable coins"), nil, nil
		}

		// Get a random IRO plan
		plans := k.GetAllPlans(ctx, true)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBuy, "no plans available"), nil, nil
		}
		plan := plans[r.Intn(len(plans))]

		// Generate random amount to buy
		amount := simtypes.RandomAmount(r, plan.TotalAllocation.Amount.Sub(plan.SoldAmt))

		// FIXME: maxAmount should be calculated based on amount
		maxAmount := math.NewInt(1000000000000).Mul(DYM)

		msg := types.MsgBuy{
			Buyer:         simAccount.Address.String(),
			PlanId:        fmt.Sprintf("%d", plan.Id),
			Amount:        amount,
			MaxCostAmount: maxAmount,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           &msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBuyExactSpend generates a MsgBuyExactSpend with random values
func SimulateMsgBuyExactSpend(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)

		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgExactSpend, "no spendable coins"), nil, nil
		}

		// Get a random IRO plan
		plans := k.GetAllPlans(ctx, true)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgExactSpend, "no plans available"), nil, nil
		}
		plan := plans[r.Intn(len(plans))]

		// Generate random amount to spend
		spendAmount := simtypes.RandomAmount(r, DYM.MulRaw(100000000))
		// FIXME: minOutTokens should be calculated based on spendAmount
		minOutTokens := math.ZeroInt()

		msg := types.MsgBuyExactSpend{
			Buyer:              simAccount.Address.String(),
			PlanId:             fmt.Sprintf("%d", plan.Id),
			Spend:              spendAmount,
			MinOutTokensAmount: minOutTokens,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           &msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgSell generates a MsgSell with random values
func SimulateMsgSell(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		account := ak.GetAccount(ctx, simAccount.Address)

		spendable := bk.SpendableCoins(ctx, account.GetAddress())
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSell, "no spendable coins"), nil, nil
		}

		// Get a random IRO plan
		plans := k.GetAllPlans(ctx, true)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSell, "no plans available"), nil, nil
		}
		plan := plans[r.Intn(len(plans))]

		// Generate random amount to sell
		sellAmount := simtypes.RandomAmount(r, spendable.AmountOf(types.IRODenom(plan.RollappId)))
		// FIXME: minIncomeAmount should be calculated based on sellAmount
		minIncomeAmount := math.ZeroInt()

		msg := types.MsgSell{
			Seller:          simAccount.Address.String(),
			PlanId:          fmt.Sprintf("%d", plan.Id),
			Amount:          sellAmount,
			MinIncomeAmount: minIncomeAmount,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           nil,
			Msg:           &msg,
			MsgType:       msg.Type(),
			Context:       ctx,
			SimAccount:    simAccount,
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
