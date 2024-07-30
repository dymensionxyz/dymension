package keeper_test

import (
	"strconv"
	"testing"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

const (
	alice = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	bob   = "dym1d0wlmz987qlurs6e3kc6zd25z6wsdmnwx8tafy"
)

type RollappTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer    types.MsgServer
	seqMsgServer sequencertypes.MsgServer
	queryClient  types.QueryClient
}

type setupOptions struct {
	accountFund sdk.Coin
	account     string
}

type setupOption func(*setupOptions)

func withAccountFund(accountFund sdk.Coin) setupOption {
	return func(options *setupOptions) {
		options.accountFund = accountFund
	}
}

func withAccount(account string) setupOption {
	return func(options *setupOptions) {
		options.account = account
	}
}

func (suite *RollappTestSuite) SetupTest(opts ...setupOption) {
	params := types.DefaultParams()
	options := setupOptions{
		accountFund: params.AliasFeeTable["3"],
		account:     alice,
	}

	for _, opt := range opts {
		opt(&options)
	}

	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	err := app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	suite.Require().NoError(err)
	err = app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	suite.Require().NoError(err)
	app.RollappKeeper.SetParams(ctx, params)

	apptesting.FundAccount(app, ctx, sdk.MustAccAddressFromBech32(options.account), sdk.NewCoins(options.accountFund))

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.msgServer = keeper.NewMsgServerImpl(*app.RollappKeeper)
	suite.seqMsgServer = sequencerkeeper.NewMsgServerImpl(app.SequencerKeeper)
	suite.Ctx = ctx
	suite.queryClient = queryClient
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RollappTestSuite))
}

func createNRollapp(keeper *keeper.Keeper, ctx sdk.Context, n int) (items []types.Rollapp, rollappSummaries []types.RollappSummary) {
	items, rollappSummaries = make([]types.Rollapp, n), make([]types.RollappSummary, n)

	for i := range items {
		items[i].RollappId = strconv.Itoa(i)
		keeper.SetRollapp(ctx, items[i])

		rollappSummaries[i] = types.RollappSummary{
			RollappId: items[i].RollappId,
		}
	}

	return
}
