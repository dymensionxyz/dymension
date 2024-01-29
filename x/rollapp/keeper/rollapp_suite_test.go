package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

var (
	rollappModuleAddress string
)

type RollappTestSuite struct {
	suite.Suite

	app         *app.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *RollappTestSuite) SetupTest(deployerWhitelist ...types.DeployerParams) {
	app := app.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.RollappKeeper.SetParams(ctx, types.NewParams(true, 2, deployerWhitelist))
	rollappModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.RollappKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RollappTestSuite))
}
