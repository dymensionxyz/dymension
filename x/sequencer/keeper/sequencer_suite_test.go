package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"

	"github.com/cometbft/cometbft/libs/rand"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
)

type SequencerTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func TestSequencerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SequencerTestSuite))
}

func (suite *SequencerTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.SequencerKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.msgServer = keeper.NewMsgServerImpl(app.SequencerKeeper)
	suite.Ctx = ctx
	suite.queryClient = queryClient
}

func (suite *SequencerTestSuite) CreateDefaultRollapp() string {
	rollapp := rollapptypes.Rollapp{
		RollappId:     rand.Str(8),
		Creator:       alice,
		Version:       0,
		MaxSequencers: 5,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)
	return rollapp.GetRollappId()
}

func (suite *SequencerTestSuite) CreateDefaultSequencer(ctx sdk.Context, rollappId string) string {
	return suite.CreateSequencerWithBond(ctx, rollappId, bond)
}

func (suite *SequencerTestSuite) CreateSequencerWithBond(ctx sdk.Context, rollappId string, bond sdk.Coin) string {
	pubkey1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pubkey1.Address())
	pkAny1, err := codectypes.NewAnyWithValue(pubkey1)
	suite.Require().Nil(err)

	// fund account
	err = bankutil.FundAccount(suite.App.BankKeeper, ctx, addr1, sdk.NewCoins(bond))
	suite.Require().Nil(err)

	sequencerMsg1 := types.MsgCreateSequencer{
		Creator:      addr1.String(),
		DymintPubKey: pkAny1,
		Bond:         bond,
		RollappId:    rollappId,
		Description:  types.Description{},
	}
	_, err = suite.msgServer.CreateSequencer(ctx, &sequencerMsg1)
	suite.Require().Nil(err)
	return addr1.String()
}
