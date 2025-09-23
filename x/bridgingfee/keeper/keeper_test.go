package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/keeper"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

// SetupTest sets up the test environment
func (s *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(s.T())
	ctx := app.NewContext(false)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, keeper.NewQueryServerImpl(app.BridgingFeeKeeper))
	queryClient := types.NewQueryClient(queryHelper)
	msgServer := keeper.NewMsgServerImpl(app.BridgingFeeKeeper)

	s.App = app
	s.Ctx = ctx
	s.queryClient = queryClient
	s.msgServer = msgServer
}

// Helper function to create test accounts
func (s *KeeperTestSuite) CreateRandomAccount() sdk.AccAddress {
	return apptesting.CreateRandomAccounts(1)[0]
}
