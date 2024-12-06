package simulation

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

const (
	WeightCreateRollapp            = 100
	WeightUpdateRollappInformation = 50
	WeightTransferOwnership        = 50
	WeightUpdateState              = 100
	WeightAddApp                   = 50
	WeightUpdateApp                = 50
	WeightRemoveApp                = 50
	WeightFraudProposal            = 20
	WeightMarkObsoleteRollapps     = 20
	WeightForceGenesisInfoChange   = 20
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

type BankKeeper interface {
	keeper.BankKeeper
}

type Keepers struct {
	Acc  AccountKeeper
	Bank BankKeeper
}

type OpFactory struct {
	keeper.Keeper
	k Keepers
	module.SimulationState
}

func NewOpFactory(k keeper.Keeper, ks Keepers, simState module.SimulationState) OpFactory {
	return OpFactory{
		Keeper:          k,
		k:               ks,
		SimulationState: simState,
	}
}

// WeightedOperations returns the simulation operations (messages) with weights.
func (f OpFactory) Messages() simulation.WeightedOperations {

	return simulation.WeightedOperations{}
}
