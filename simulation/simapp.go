package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp/helpers"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/simulation/types"
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

// GenAndDeliverMsgWithRandFees generates a transaction with a random fee and expected Error flag (bExpectedError).
// GenAndDeliverMsgWithRandFees wraps GenAndDeliverTxWithRand Fees and checks whether or not the operation
// failed as expected by bExpectedError flag
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

	operationMsg, additionalFutureOperation, err := GenAndDeliverTxWithRandFees(txCtx)

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

// GenAndDeliverTxWithRandFees generates a transaction with a random fee and delivers it.
// Copied from github.com/cosmos/cosmos-sdk/x/simulation/util because of the need to increase the gas
// as haedcoded passed in helpers.GenTx
func GenAndDeliverTxWithRandFees(txCtx simulation.OperationInput) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	spendable := txCtx.Bankkeeper.SpendableCoins(txCtx.Context, account.GetAddress())

	var fees sdk.Coins
	var err error

	coins, hasNeg := spendable.SafeSub(txCtx.CoinsSpentInMsg)
	if hasNeg {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "message doesn't leave room for fees"), nil, err
	}

	fees, err = simtypes.RandomFees(txCtx.R, txCtx.Context, coins)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate fees"), nil, err
	}
	return GenAndDeliverTx(txCtx, fees)
}

// GenAndDeliverTx generates a transactions and delivers it.
// Copied from github.com/cosmos/cosmos-sdk/x/simulation/util
func GenAndDeliverTx(txCtx simulation.OperationInput, fees sdk.Coins) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	account := txCtx.AccountKeeper.GetAccount(txCtx.Context, txCtx.SimAccount.Address)
	tx, err := helpers.GenTx(
		txCtx.TxGen,
		[]sdk.Msg{txCtx.Msg},
		fees,
		10*helpers.DefaultGenTxGas,
		txCtx.Context.ChainID(),
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		txCtx.SimAccount.PrivKey,
	)

	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to generate mock tx"), nil, err
	}

	_, _, err = txCtx.App.Deliver(txCtx.TxGen.TxEncoder(), tx)
	if err != nil {
		return simtypes.NoOpMsg(txCtx.ModuleName, txCtx.MsgType, "unable to deliver tx"), nil, err
	}

	return simtypes.NewOperationMsg(txCtx.Msg, true, "", txCtx.Cdc), nil, nil

}
