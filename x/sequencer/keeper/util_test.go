package keeper_test

import (
	"reflect"
	"strconv"
	"testing"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

const (
	alice        = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bech32Prefix = "eth"
)

var bond = types.DefaultParams().MinBond

// Prevent strconv unused error
var _ = strconv.IntSize

func createNSequencer(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.Sequencer {
	items := make([]types.Sequencer, n)
	for i := range items {
		seq := types.Sequencer{
			Address: strconv.Itoa(i),
			Status:  types.Bonded,
		}
		items[i] = seq

		keeper.SetSequencer(ctx, items[i])
	}
	return items
}

type SequencerTestSuite struct {
	apptesting.KeeperTestHelper
	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func TestSequencerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(SequencerTestSuite))
}

func (s *SequencerTestSuite) k() keeper.Keeper {
	return s.App.SequencerKeeper
}

func (s *SequencerTestSuite) SetupTest() {
	app := apptesting.Setup(s.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.SequencerKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	s.App = app
	s.msgServer = keeper.NewMsgServerImpl(app.SequencerKeeper)
	s.Ctx = ctx
	s.queryClient = queryClient
}

func (s *SequencerTestSuite) createRollapp() (string, cryptotypes.PubKey) {
	pubkey := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	return s.createRollappWithInitialSeq(addr.String()), pubkey
}

func (s *SequencerTestSuite) createRollappWithInitialSeq(initSeq string) string {
	rollapp := rollapptypes.Rollapp{
		RollappId: urand.RollappID(),
		Owner:     sample.AccAddress(),
		GenesisInfo: rollapptypes.GenesisInfo{
			Bech32Prefix:    "rol",
			GenesisChecksum: "checksum",
			NativeDenom:     rollapptypes.DenomMetadata{Display: "DEN", Base: "aden", Exponent: 18},
			InitialSupply:   sdk.NewInt(1000),
		},
		InitialSequencer: initSeq,
	}
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)
	return rollapp.GetRollappId()
}

func (s *SequencerTestSuite) createSequencer(ctx sdk.Context, rollappId string, pk cryptotypes.PubKey) string {
	return s.createSequencerWithBond(ctx, rollappId, pk, bond)
}

func (s *SequencerTestSuite) createSequencerWithBond(ctx sdk.Context, rollappId string, pk cryptotypes.PubKey, bond sdk.Coin) string {
	pkAny, err := codectypes.NewAnyWithValue(pk)
	s.Require().Nil(err)

	addr := sdk.AccAddress(pk.Address())
	// fund account
	err = bankutil.FundAccount(s.App.BankKeeper, ctx, addr, sdk.NewCoins(bond))
	s.Require().Nil(err)

	sequencerMsg1 := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types.SequencerMetadata{
			Rpcs:    []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
			EvmRpcs: []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
		},
	}
	_, err = s.msgServer.CreateSequencer(ctx, &sequencerMsg1)
	s.Require().NoError(err)
	return addr.String()
}

// ---------------------------------------
// verifyAll receives a list of expected results and a map of sequencerAddress->sequencer
// the function verifies that the map contains all the sequencers that are in the list and only them
func (s *SequencerTestSuite) verifyAll(sequencersExpect []*types.Sequencer, sequencersRes map[string]*types.Sequencer) {
	// check number of items are equal
	s.Require().EqualValues(len(sequencersExpect), len(sequencersRes))
	for i := 0; i < len(sequencersExpect); i++ {
		sequencerExpect := sequencersExpect[i]
		sequencerRes := sequencersRes[sequencerExpect.GetAddress()]
		s.equalSequencer(sequencerExpect, sequencerRes)
	}
}

// getAll quires for all existing sequencers and returns a map of sequencerId->sequencer
func getAll(suite *SequencerTestSuite) (map[string]*types.Sequencer, int) {
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	totalChecked := 0
	totalRes := 0
	nextKey := []byte{}
	sequencersRes := make(map[string]*types.Sequencer)
	for {
		queryAllResponse, err := suite.queryClient.Sequencers(goCtx,
			&types.QuerySequencersRequest{
				Pagination: &query.PageRequest{
					Key:        nextKey,
					Offset:     0,
					Limit:      0,
					CountTotal: true,
					Reverse:    false,
				},
			})
		suite.Require().Nil(err)

		if totalRes == 0 {
			totalRes = int(queryAllResponse.GetPagination().GetTotal())
		}

		for i := 0; i < len(queryAllResponse.Sequencers); i++ {
			sequencerRes := queryAllResponse.Sequencers[i]
			sequencersRes[sequencerRes.GetAddress()] = &sequencerRes
		}
		totalChecked += len(queryAllResponse.Sequencers)
		nextKey = queryAllResponse.GetPagination().GetNextKey()

		if nextKey == nil {
			break
		}
	}

	return sequencersRes, totalRes
}

// equalSequencer receives two sequencers and compares them. If they are not equal, fails the test
func (s *SequencerTestSuite) equalSequencer(s1 *types.Sequencer, s2 *types.Sequencer) {
	eq := compareSequencers(s1, s2)
	s.Require().True(eq, "expected: %+v\nfound: %+v", *s1, *s2)
}

func compareSequencers(s1, s2 *types.Sequencer) bool {
	if s1.Address != s2.Address {
		return false
	}

	s1Pubkey := s1.DymintPubKey
	s2Pubkey := s2.DymintPubKey
	if !s1Pubkey.Equal(s2Pubkey) {
		return false
	}
	if s1.RollappId != s2.RollappId {
		return false
	}

	if s1.Status != s2.Status {
		return false
	}

	if !s1.Tokens.IsEqual(s2.Tokens) {
		return false
	}

	if !s1.NoticePeriodTime.Equal(s2.NoticePeriodTime) {
		return false
	}
	if !reflect.DeepEqual(s1.Metadata, s2.Metadata) {
		return false
	}
	return true
}