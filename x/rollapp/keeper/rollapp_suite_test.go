package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

const (
	transferEventCount            = 3 // As emitted by the bank
	createEventCount              = 8
	playEventCountFirst           = 8 // Extra "sender" attribute emitted by the bank
	playEventCountNext            = 7
	rejectEventCount              = 4
	rejectEventCountWithTransfer  = 5 // Extra "sender" attribute emitted by the bank
	forfeitEventCount             = 4
	forfeitEventCountWithTransfer = 5 // Extra "sender" attribute emitted by the bank
	alice                         = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	bob                           = "dym1d0wlmz987qlurs6e3kc6zd25z6wsdmnwx8tafy"
	carol                         = "dym1zka35fkgdxmhl8ncjgwkrah0e8kpvd8vkn0vsp"
	balAlice                      = 50000000
	balBob                        = 20000000
	balCarol                      = 10000000
	foreignToken                  = "foreignToken"
	balTokenAlice                 = 5
	balTokenBob                   = 2
	balTokenCarol                 = 1
)

var rollappModuleAddress string

type RollappTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func (suite *RollappTestSuite) SetupTest(deployerWhitelist ...types.DeployerParams) {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.RollappKeeper.SetParams(ctx, types.NewParams(true, 2, deployerWhitelist))
	rollappModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.msgServer = keeper.NewMsgServerImpl(*app.RollappKeeper)
	suite.Ctx = ctx
	suite.queryClient = queryClient
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RollappTestSuite))
}

func createNRollapp(keeper *keeper.Keeper, ctx sdk.Context, n int) ([]types.Rollapp, []types.RollappSummary) {
	items := make([]types.Rollapp, n)
	for i := range items {
		items[i].RollappId = strconv.Itoa(i)
		keeper.SetRollapp(ctx, items[i])
	}

	rollappSummaries := []types.RollappSummary{}
	for _, item := range items {
		rollappSummary := types.RollappSummary{
			RollappId: item.RollappId,
		}
		rollappSummaries = append(rollappSummaries, rollappSummary)
	}

	return items, rollappSummaries
}
