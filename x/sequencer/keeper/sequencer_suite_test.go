package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app"
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

	"github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type SequencerTestSuite struct {
	suite.Suite

	sequencerModuleAddress string

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

	suite.sequencerModuleAddress = app.AccountKeeper.GetModuleAddress(types.ModuleName).String()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.SequencerKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.msgServer = keeper.NewMsgServerImpl(app.SequencerKeeper)
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *SequencerTestSuite) CreateDefaultRollapp() string {
	rollapp := rollapptypes.Rollapp{
		RollappId:     rand.Str(8),
		Creator:       alice,
		Version:       0,
		MaxSequencers: 5,
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)
	return rollapp.GetRollappId()
}

func (suite *SequencerTestSuite) CreateDefaultSequencer(ctx sdk.Context, rollappId string) string {
	pubkey1 := secp256k1.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pubkey1.Address())
	pkAny1, err := codectypes.NewAnyWithValue(pubkey1)
	suite.Require().Nil(err)

	//fund account
	err = bankutil.FundAccount(suite.app.BankKeeper, ctx, addr1, sdk.NewCoins(bond))
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
