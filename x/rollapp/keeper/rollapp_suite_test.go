package keeper_test

import (
	_ "strconv"
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

const (
	alice           = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	bob             = "dym1d0wlmz987qlurs6e3kc6zd25z6wsdmnwx8tafy"
	registrationFee = "1000000000000000000adym"
	hubChainID      = "dymension_100-1"
)

type RollappTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer    types.MsgServer
	seqMsgServer sequencertypes.MsgServer
	queryClient  types.QueryClient
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RollappTestSuite))
}

func (s *RollappTestSuite) SetupTest() {
	app := apptesting.Setup(s.T())
	s.App = app
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{ChainID: hubChainID})

	err := app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	s.Require().NoError(err)
	err = app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	s.Require().NoError(err)
	regFee, _ := sdk.ParseCoinNormalized(registrationFee)
	s.k().SetParams(ctx, types.DefaultParams().WithDisputePeriodInBlocks(2))

	aliceBal := sdk.NewCoins(regFee.AddAmount(regFee.Amount.Mul(sdk.NewInt(50))))
	apptesting.FundAccount(app, ctx, sdk.MustAccAddressFromBech32(alice), aliceBal)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, s.k())
	queryClient := types.NewQueryClient(queryHelper)

	s.msgServer = keeper.NewMsgServerImpl(s.k())
	s.seqMsgServer = sequencerkeeper.NewMsgServerImpl(app.SequencerKeeper)
	s.Ctx = ctx
	s.queryClient = queryClient
}

func (s *RollappTestSuite) k() *keeper.Keeper {
	return s.App.RollappKeeper
}

func (s *RollappTestSuite) assertNotForked(rollappID string) {
	rollapp, _ := s.k().GetRollapp(s.Ctx, rollappID)
	s.Zero(rollapp.LatestRevision().Number)
}

func (s *RollappTestSuite) GetRollappLastHeight(rollappID string) uint64 {
	stateInfo, ok := s.k().GetLatestStateInfo(s.Ctx, rollappID)
	s.Require().True(ok)
	return stateInfo.GetLatestHeight()
}
