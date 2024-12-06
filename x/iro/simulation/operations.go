package simulation

import (
	"fmt"
	"math/rand"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/iro/keeper"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	WeightFundModule = 100
	WeightCreatePlan = 100
	WeightBuy        = 100
	WeightSell       = 100
	WeightClaim      = 50
)

type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	types.BankKeeper
}

type AccountKeeper interface {
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

// Messages returns all the simulation operations for the IRO module
func (f OpFactory) Messages() []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{
		//simulation.NewWeightedOperation(

		simulation.NewWeightedOperation(
			WeightCreatePlan,
			f.CreatePlanOp,
		),
		simulation.NewWeightedOperation(
			WeightBuy,
			f.BuyOp,
		),
		simulation.NewWeightedOperation(
			WeightSell,
			f.SellOp,
		),
		simulation.NewWeightedOperation(
			WeightClaim,
			f.ClaimOp,
		),
	}
}

func (f OpFactory) Proposals() []simtypes.WeightedProposalContent {
	return []simtypes.WeightedProposalContent{}
}

// SimulateMsgCreatePlan simulates creating an IRO plan.
func SimulateMsgCreatePlan(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	rk types.RollappKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		owner, _ := simtypes.RandomAcc(r, accs)

		allocation, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1_000_000), sdk.NewInt(10_000_000))
		startTime := ctx.BlockTime().Add(simtypes.RandDuration(r, 1*time.Hour))
		duration := simtypes.RandDuration(r, 2*time.Hour)

		rollapps := rk.GetAllRollaps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreatePlan, "no rollapps"), nil, nil
		}
		rollapp := dymsimtypes.RandChoice(r, rollapps)

		curve := types.NewBondingCurve(
			sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 3), // small M
			sdk.NewDec(1),                                                   // N = 1
			sdk.ZeroDec(),                                                   // C = 0 for simplicity
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
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgBuy simulates buying tokens from an IRO plan.
func SimulateMsgBuy(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	rk types.RollappKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := k.GetAllPlans(ctx, false)
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
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgSell simulates selling tokens back to the IRO plan.
func SimulateMsgSell(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	rk types.RollappKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := k.GetAllPlans(ctx, false)
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
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgClaim simulates claiming tokens after the plan is settled.
func SimulateMsgClaim(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	rk types.RollappKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		plans := k.GetAllPlans(ctx, false)
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
			AccountKeeper: ak,
			Bankkeeper:    bk,
			ModuleName:    types.ModuleName,
		}

		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}
