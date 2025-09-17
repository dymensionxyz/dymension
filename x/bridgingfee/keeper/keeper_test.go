package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
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

// CreateTokenWithPool creates a token in the warp module and a corresponding pool for fee calculations
func (s *KeeperTestSuite) CreateTokenWithPool(originDenom string, amount math.Int) string {
	s.T().Helper()

	// Fund the account with the origin denom
	coins := sdk.NewCoins(sdk.NewCoin(originDenom, amount))
	s.FundAcc(sdk.MustAccAddressFromBech32(apptesting.Alice), coins)

	// Create pool between origin denom and base denom (adym)
	baseDenom := "adym"
	poolCoins := sdk.NewCoins(
		sdk.NewCoin(originDenom, amount.QuoRaw(2)), // half for the pool
		sdk.NewCoin(baseDenom, amount.QuoRaw(2)),   // half for the pool
	)
	s.PreparePoolWithCoins(poolCoins)

	// TODO: This would normally create a token in the warp module
	// For now, we'll return a dummy token ID that represents this token
	return "0x" + originDenom + "000000000000000000000000000000000000"
}

// Helper function to create test accounts
func (s *KeeperTestSuite) CreateRandomAccount() sdk.AccAddress {
	return apptesting.CreateRandomAccounts(1)[0]
}