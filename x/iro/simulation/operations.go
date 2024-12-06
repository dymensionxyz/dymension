package simulation

import (
	"fmt"
	"math/rand"
	"time"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
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

// CreatePlanOp tries to create a new IRO plan with random parameters.
func (f OpFactory) CreatePlanOp(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, id string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	// Choose a random account as owner
	owner, _ := simtypes.RandomAcc(r, accs)

	// Random allocation and times
	allocation, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1_000_000), sdk.NewInt(10_000_000))
	startTime := ctx.BlockTime().Add(simtypes.RandDuration(r, 1*time.Hour))
	duration := simtypes.RandDuration(r, 2*time.Hour)

	rollappList := f.k.Rollapp.GetAllRollapp(ctx)
	if len(rollappList) == 0 {
		return simtypes.NoOpMsg(types.ModuleName, "create_plan", "no rollapps"), nil, nil
	}
	rollapp := dymsimtypes.RandChoice(r, rollappList)

	curve := types.NewBondingCurve(
		sdk.NewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10)), 3), // small M
		sdk.NewDec(1),                                                   // N = 1
		sdk.ZeroDec(),                                                   // C = 0 for simplicity
	)

	// Default incentives
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
		R:               r,
		App:             app,
		TxGen:           f.SimulationState.TxConfig,
		Ctx:             ctx,
		Msg:             msg,
		Accounts:        accs,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: sdk.Coins{},
	}

	return simulation.GenAndDeliverTxWithRandFees(txCtx)
}

// BuyOp attempts to buy some tokens from an existing IRO plan.
func (f OpFactory) BuyOp(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, id string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	plans := f.Keeper.GetAllPlans(ctx, false)
	if len(plans) == 0 {
		return simtypes.NoOpMsg(types.ModuleName, "buy", "no plans"), nil, nil
	}

	plan := dymsimtypes.RandChoice(r, plans)
	if plan.IsSettled() || ctx.BlockTime().Before(plan.StartTime) {
		return simtypes.NoOpMsg(types.ModuleName, "buy", "plan not tradable"), nil, nil
	}

	// Choose random buyer
	buyer, _ := simtypes.RandomAcc(r, accs)
	amount, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1), sdk.NewInt(1000))

	msg := &types.MsgBuy{
		Buyer:         buyer.Address.String(),
		PlanId:        fmt.Sprintf("%d", plan.Id),
		Amount:        amount,
		MaxCostAmount: amount.MulRaw(2), // arbitrary margin
	}

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           f.SimulationState.TxConfig,
		Ctx:             ctx,
		Msg:             msg,
		Accounts:        accs,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: sdk.Coins{}, // buyer pays for tokens
	}

	return simulation.GenAndDeliverTxWithRandFees(txCtx)
}

// SellOp attempts to sell some tokens from the buyer back to the plan.
func (f OpFactory) SellOp(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, id string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	plans := f.Keeper.GetAllPlans(ctx, false)
	if len(plans) == 0 {
		return simtypes.NoOpMsg(types.ModuleName, "sell", "no plans"), nil, nil
	}

	plan := dymsimtypes.RandChoice(r, plans)
	if plan.IsSettled() || ctx.BlockTime().Before(plan.StartTime) {
		return simtypes.NoOpMsg(types.ModuleName, "sell", "plan not tradable"), nil, nil
	}

	seller, _ := simtypes.RandomAcc(r, accs)
	amount, _ := dymsimtypes.RandIntBetween(r, sdk.NewInt(1), sdk.NewInt(1000))

	msg := &types.MsgSell{
		Seller:          seller.Address.String(),
		PlanId:          fmt.Sprintf("%d", plan.Id),
		Amount:          amount,
		MinIncomeAmount: amount.QuoRaw(2), // arbitrary expectation
	}

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           f.SimulationState.TxConfig,
		Ctx:             ctx,
		Msg:             msg,
		Accounts:        accs,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: sdk.Coins{}, // seller must have FUT tokens
	}

	return simulation.GenAndDeliverTxWithRandFees(txCtx)
}

// ClaimOp attempts to claim tokens after the plan is settled.
func (f OpFactory) ClaimOp(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, id string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	plans := f.Keeper.GetAllPlans(ctx, false)
	if len(plans) == 0 {
		return simtypes.NoOpMsg(types.ModuleName, "claim", "no plans"), nil, nil
	}

	plan := dymsimtypes.RandChoice(r, plans)
	if !plan.IsSettled() {
		return simtypes.NoOpMsg(types.ModuleName, "claim", "plan not settled"), nil, nil
	}

	claimer, _ := simtypes.RandomAcc(r, accs)
	msg := &types.MsgClaim{
		Claimer: claimer.Address.String(),
		PlanId:  fmt.Sprintf("%d", plan.Id),
	}

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           f.SimulationState.TxConfig,
		Ctx:             ctx,
		Msg:             msg,
		Accounts:        accs,
		ModuleName:      types.ModuleName,
		CoinsSpentInMsg: nil,
	}

	return simulation.GenAndDeliverTxWithRandFees(txCtx)
}
