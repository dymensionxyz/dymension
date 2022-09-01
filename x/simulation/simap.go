package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/x/simulation/types"
)

// GlobalRollappIdList is a list of created rollapps
var GlobalRollappList []types.SimRollapp = []types.SimRollapp{}

// GlobalSequencerAddressesList is a list of created sequencers
var GlobalSequencerAddressesList []types.SimSequencer = []types.SimSequencer{}

// FindAccount find a specific address from an account list
func FindAccount(accs []simtypes.Account, address string) (simtypes.Account, bool) {
	creator, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return simtypes.FindAccount(accs, creator)
}

// RandomRollapp picks and returns a random rollapp from an array and returs its
// position in the array.
func RandomRollapp(r *rand.Rand, rollappList []types.SimRollapp) (types.SimRollapp, int) {
	idx := r.Intn(len(rollappList))
	return rollappList[idx], idx
}

// RandomSequencer picks and returns a random sequencer from an array and returs its
// position in the array.
func RandomSequencer(r *rand.Rand, sequencerList []types.SimSequencer) (types.SimSequencer, int) {
	idx := r.Intn(len(sequencerList))
	return sequencerList[idx], idx
}

func GenAndDeliverMsgWithRandFees(
	msg sdk.Msg,
	msgType string,
	moduleName string,
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx *sdk.Context,
	simAccount *simtypes.Account,
	bk types.BankKeeper,
	ak types.AccountKeeper,
	futureOperation []simtypes.FutureOperation,
	bExpectedError bool) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

	spendableCoins := bk.SpendableCoins(*ctx, simAccount.Address)

	if spendableCoins.Empty() {
		return simtypes.NoOpMsg(moduleName, msgType, "unable to grant empty coins as SpendLimit"), nil, nil
	}

	txCtx := simulation.OperationInput{
		R:               r,
		App:             app,
		TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
		Cdc:             nil,
		Msg:             msg,
		MsgType:         msgType,
		CoinsSpentInMsg: spendableCoins,
		Context:         *ctx,
		SimAccount:      *simAccount,
		AccountKeeper:   ak,
		Bankkeeper:      bk,
		ModuleName:      moduleName,
	}

	operationMsg, additionalFutureOperation, err := simulation.GenAndDeliverTxWithRandFees(txCtx)

	futureOperation = append(futureOperation, additionalFutureOperation...)
	if bExpectedError {
		if err == nil {
			panic(err)
		}
		err = nil
	} else {
		if err != nil {
			panic(err)
		}
	}
	return operationMsg, futureOperation, err
}
