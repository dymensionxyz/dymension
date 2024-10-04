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
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

const (
	alice           = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	bob             = "dym1d0wlmz987qlurs6e3kc6zd25z6wsdmnwx8tafy"
	registrationFee = "1000000000000000000adym"
)

type RollappTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer    types.MsgServer
	seqMsgServer sequencertypes.MsgServer
	queryClient  types.QueryClient
}

func (suite *RollappTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	err := app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	suite.Require().NoError(err)
	err = app.BankKeeper.SetParams(ctx, banktypes.DefaultParams())
	suite.Require().NoError(err)
	regFee, _ := sdk.ParseCoinNormalized(registrationFee)
	app.RollappKeeper.SetParams(ctx, types.DefaultParams().WithDisputePeriodInBlocks(2))

	aliceBal := sdk.NewCoins(regFee.AddAmount(regFee.Amount.Mul(sdk.NewInt(50))))
	apptesting.FundAccount(app, ctx, sdk.MustAccAddressFromBech32(alice), aliceBal)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.RollappKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.msgServer = keeper.NewMsgServerImpl(*app.RollappKeeper)
	suite.seqMsgServer = sequencerkeeper.NewMsgServerImpl(app.SequencerKeeper)
	suite.Ctx = ctx
	suite.queryClient = queryClient
}

func (suite *RollappTestSuite) keeper() *keeper.Keeper {
	return suite.App.RollappKeeper
}

func (suite *RollappTestSuite) nextBlock() {
	h := suite.Ctx.BlockHeight()
	suite.Ctx = suite.Ctx.WithBlockHeight(h + 1)
}

func TestRollappKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(RollappTestSuite))
}

func (suite *RollappTestSuite) IsRollappVulnerable(rollappID string) bool {
	ra, ok := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappID)
	suite.Require().True(ok)
	return ra.IsVulnerable()
}

func (suite *RollappTestSuite) GetRollappLastHeight(rollappID string) uint64 {
	stateInfo, ok := suite.App.RollappKeeper.GetLatestStateInfo(suite.Ctx, rollappID)
	suite.Require().True(ok)
	return stateInfo.GetLatestHeight() + 1
}

func (suite *RollappTestSuite) TestIsRollAppExists() {
	suite.Run("pass - rollapp does not exists", func() {
		gotExists := suite.App.RollappKeeper.IsRollAppExists(suite.Ctx, "void")
		suite.Require().False(gotExists)
	})

	suite.Run("pass - rollapp exists", func() {
		const rollAppId = "rollapp_9630-9630"
		suite.Require().False(suite.App.RollappKeeper.IsRollAppExists(suite.Ctx, rollAppId))

		goCtx := sdk.WrapSDKContext(suite.Ctx)
		rollappMsg := types.MsgCreateRollapp{
			Creator:          alice,
			RollappId:        rollAppId,
			InitialSequencer: sample.AccAddress(),
			Alias:            "rollapp9630",
			VmType:           types.Rollapp_EVM,
			Metadata:         &mockRollappMetadata,
			GenesisInfo:      mockGenesisInfo,
		}

		suite.FundForAliasRegistration(rollappMsg)

		_, err := suite.msgServer.CreateRollapp(goCtx, &rollappMsg)
		suite.Require().NoError(err)

		gotExists := suite.App.RollappKeeper.IsRollAppExists(suite.Ctx, rollAppId)
		suite.Require().True(gotExists)
	})
}

func (suite *RollappTestSuite) TestGetRollAppIdByEIP155() {
	const rollAppId = "rollapp_9631-9630"
	const rollAppEip155Id uint64 = 9631
	suite.Run("pass - when rollapp does not exists", func() {
		_, gotFound := suite.App.RollappKeeper.GetRollAppIdByEIP155(suite.Ctx, rollAppEip155Id)
		suite.Require().False(gotFound)
	})

	suite.Run("pass - rollapp exists", func() {
		_, gotFound := suite.App.RollappKeeper.GetRollAppIdByEIP155(suite.Ctx, rollAppEip155Id)
		suite.Require().False(gotFound)

		goCtx := sdk.WrapSDKContext(suite.Ctx)
		rollappMsg := types.MsgCreateRollapp{
			Creator:          alice,
			RollappId:        rollAppId,
			InitialSequencer: sample.AccAddress(),
			Alias:            "rollapp9630",
			VmType:           types.Rollapp_EVM,
			Metadata:         &mockRollappMetadata,
			GenesisInfo:      mockGenesisInfo,
		}

		suite.FundForAliasRegistration(rollappMsg)

		_, err := suite.msgServer.CreateRollapp(goCtx, &rollappMsg)
		suite.Require().NoError(err)

		gotRollAppId, gotFound := suite.App.RollappKeeper.GetRollAppIdByEIP155(suite.Ctx, rollAppEip155Id)
		suite.Require().True(gotFound)
		suite.Require().Equal(rollAppId, gotRollAppId)
	})
}
