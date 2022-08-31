package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
)

// list of created rollapps
var globalRollappIdList []string = []string{}

// FindAccount find a specific address from an account list
func FindAccount(accs []simtypes.Account, address string) (simtypes.Account, bool) {
	creator, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		panic(err)
	}
	return simtypes.FindAccount(accs, creator)
}

func GenAndDeliverMsgWithRandFees(
	msg sdk.Msg,
	msgType string,
	r *rand.Rand,
	app *baseapp.BaseApp,
	ctx *sdk.Context,
	simAccount *simtypes.Account,
	bk *types.BankKeeper,
	ak *types.AccountKeeper,
	futureOperation []simtypes.FutureOperation,
	bExpectedError bool) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {

	spendableCoins := (*bk).SpendableCoins(*ctx, simAccount.Address)

	if spendableCoins.Empty() {
		return simtypes.NoOpMsg(types.ModuleName, msgType, "unable to grant empty coins as SpendLimit"), nil, nil
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
		AccountKeeper:   *ak,
		Bankkeeper:      *bk,
		ModuleName:      types.ModuleName,
	}

	operationMsg, additionalFutureOperation, err := simulation.GenAndDeliverTxWithRandFees(txCtx)

	futureOperation = append(futureOperation, additionalFutureOperation...)
	if bExpectedError {
		err = nil
	} else {
		if err != nil {
			panic(err)
		}
	}
	return operationMsg, futureOperation, err
}
