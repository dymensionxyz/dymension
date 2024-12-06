package simulation

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	WeightCreatePlan = 100
	WeightBuy        = 100
	WeightSell       = 100
	WeightClaim      = 50
)

type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	types.BankKeeper
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	types.AccountKeeper
}

type RollappKeeper interface {
	GetAllRollapps(ctx sdk.Context) []rollapptypes.Rollapp
	types.RollappKeeper
}

type Keepers struct {
	Bank    BankKeeper
	Acc     AccountKeeper
	Rollapp RollappKeeper
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

func (f OpFactory) Proposals() []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{}
}

// WeightedOperations returns all the operations from the IRO module with their respective weights.
func (f OpFactory) Messages() simulation.WeightedOperations {
	var weightCreatePlan, weightBuy, weightSell, weightClaim int

	f.AppParams.GetOrGenerate(
		f.Cdc, "create_plan", &weightCreatePlan, nil,
		func(_ *rand.Rand) { weightCreatePlan = WeightCreatePlan },
	)

	f.AppParams.GetOrGenerate(
		f.Cdc, "buy", &weightBuy, nil,
		func(_ *rand.Rand) { weightBuy = WeightBuy },
	)

	f.AppParams.GetOrGenerate(
		f.Cdc, "sell", &weightSell, nil,
		func(_ *rand.Rand) { weightSell = WeightSell },
	)

	f.AppParams.GetOrGenerate(
		f.Cdc, "claim", &weightClaim, nil,
		func(_ *rand.Rand) { weightClaim = WeightClaim },
	)

	protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightCreatePlan,
			f.simulateMsgCreatePlan(protoCdc),
		),
		simulation.NewWeightedOperation(
			weightBuy,
			f.simulateMsgBuy(protoCdc),
		),
		simulation.NewWeightedOperation(
			weightSell,
			f.simulateMsgSell(protoCdc),
		),
		simulation.NewWeightedOperation(
			weightClaim,
			f.simulateMsgClaim(protoCdc),
		),
	}
}

// simulateMsgCreatePlan simulates creating an IRO plan.
func (f OpFactory) simulateMsgCreatePlan(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		owner, _ := simtypes.RandomAcc(r, accs)

		allocation, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1_000_000), sdk.NewInt(10_000_000))
		startTime := ctx.BlockTime().Add(dymsimtypes.RandDuration(r, 1*time.Hour))
		duration := dymsimtypes.RandDuration(r, 2*time.Hour)

		rollapps := f.k.Rollapp.GetAllRollapps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePlan, "no rollapps"), nil, nil
		}
		rollapp := dymsimtypes.RandChoice(r, rollapps)

		curve := types.NewBondingCurve(
			sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 3), // small M
			sdk.NewDec(1),
			sdk.ZeroDec(),
		)
		incentives := types.DefaultIncentivePlanParams()

		msg := &types.MsgCreatePlan{
			Owner:               owner.Address.String(),
			AllocatedAmount:     allocation,
			StartTime:           startTime,
			IroPlanDuration:     duration,
			RollappId:           rollapp.RollappId,
			BondingCurve:        curve,
			IncentivePlanParams: incentives,
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           cdc,
			Msg:           msg,
			MsgType:       msg.Type(),
			SimAccount:    owner,
			Context:       ctx,
			AccountKeeper: f.k.Acc,
			Bankkeeper:    f.k.Bank,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// simulateMsgBuy simulates buying tokens from an IRO plan.
func (f OpFactory) simulateMsgBuy(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := f.GetAllPlans(ctx, false)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBuy, "no plans"), nil, nil
		}
		plan := dymsimtypes.RandChoice(r, plans)
		if plan.IsSettled() || ctx.BlockTime().Before(plan.StartTime) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgBuy, "plan not tradable"), nil, nil
		}

		buyer, _ := simtypes.RandomAcc(r, accs)
		amount, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1), sdk.NewInt(1000))

		msg := &types.MsgBuy{
			Buyer:         buyer.Address.String(),
			PlanId:        fmt.Sprintf("%d", plan.Id),
			Amount:        amount,
			MaxCostAmount: amount.MulRaw(2),
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           cdc,
			Msg:           msg,
			MsgType:       msg.Type(),
			SimAccount:    buyer,
			Context:       ctx,
			AccountKeeper: f.k.Acc,
			Bankkeeper:    f.k.Bank,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// simulateMsgSell simulates selling tokens back to the IRO plan.
func (f OpFactory) simulateMsgSell(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := f.GetAllPlans(ctx, false)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSell, "no plans"), nil, nil
		}

		plan := dymsimtypes.RandChoice(r, plans)
		if plan.IsSettled() || ctx.BlockTime().Before(plan.StartTime) {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgSell, "plan not tradable"), nil, nil
		}

		seller, _ := simtypes.RandomAcc(r, accs)
		amount, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1), sdk.NewInt(1000))

		msg := &types.MsgSell{
			Seller:          seller.Address.String(),
			PlanId:          fmt.Sprintf("%d", plan.Id),
			Amount:          amount,
			MinIncomeAmount: amount.QuoRaw(2),
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           cdc,
			Msg:           msg,
			MsgType:       msg.Type(),
			SimAccount:    seller,
			Context:       ctx,
			AccountKeeper: f.k.Acc,
			Bankkeeper:    f.k.Bank,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// simulateMsgClaim simulates claiming tokens after the plan is settled.
func (f OpFactory) simulateMsgClaim(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := f.GetAllPlans(ctx, false)
		if len(plans) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaim, "no plans"), nil, nil
		}

		plan := dymsimtypes.RandChoice(r, plans)
		if !plan.IsSettled() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgClaim, "plan not settled"), nil, nil
		}

		claimer, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgClaim{
			Claimer: claimer.Address.String(),
			PlanId:  fmt.Sprintf("%d", plan.Id),
		}

		txCtx := simulation.OperationInput{
			R:             r,
			App:           app,
			TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
			Cdc:           cdc,
			Msg:           msg,
			MsgType:       msg.Type(),
			SimAccount:    claimer,
			Context:       ctx,
			AccountKeeper: f.k.Acc,
			Bankkeeper:    f.k.Bank,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
