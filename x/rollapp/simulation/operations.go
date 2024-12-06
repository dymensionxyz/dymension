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
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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

type Keepers struct {
	Bank BankKeeper
	Acc  AccountKeeper
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
