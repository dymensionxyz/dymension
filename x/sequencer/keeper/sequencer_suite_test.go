package keeper_test

import (
	"testing"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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

func (suite *SequencerTestSuite) CreateDefaultRollapp() (string, cryptotypes.PubKey) {
	pubkey := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	return suite.CreateRollappWithInitialSequencer(addr.String()), pubkey
}

func (suite *SequencerTestSuite) CreateRollappWithInitialSequencer(initSeq string) string {
	rollapp := rollapptypes.Rollapp{
		RollappId:        urand.RollappID(),
		Owner:            sample.AccAddress(),
		GenesisChecksum:  "checksum",
		InitialSequencer: initSeq,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)
	return rollapp.GetRollappId()
}

func (suite *SequencerTestSuite) CreateDefaultSequencer(ctx sdk.Context, rollappId string, pk cryptotypes.PubKey) string {
	return suite.CreateSequencerWithBond(ctx, rollappId, bond, pk)
}

func (suite *SequencerTestSuite) CreateSequencerWithBond(ctx sdk.Context, rollappId string, bond sdk.Coin, pk cryptotypes.PubKey) string {
	pkAny, err := codectypes.NewAnyWithValue(pk)
	suite.Require().Nil(err)

	addr := sdk.AccAddress(pk.Address())
	// fund account
	err = bankutil.FundAccount(suite.App.BankKeeper, ctx, addr, sdk.NewCoins(bond))
	suite.Require().Nil(err)

	sequencerMsg1 := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}
	_, err = suite.msgServer.CreateSequencer(ctx, &sequencerMsg1)
	suite.Require().NoError(err)
	return addr.String()
}
