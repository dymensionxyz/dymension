package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type SequencerTestSuite struct {
	suite.Suite

	app         *app.App
	msgServer   types.MsgServer
	ctx         sdk.Context
	queryClient types.QueryClient
}

func TestSequencerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SequencerTestSuite))
}

func (suite *SequencerTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	app.RollappKeeper.SetParams(ctx, rollapptypes.DefaultParams())
	sequencerModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.SequencerKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.SequencerKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}
