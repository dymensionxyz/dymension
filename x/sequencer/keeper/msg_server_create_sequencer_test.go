package keeper_test

import (
	"fmt"
	"reflect"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/dymensionxyz/sdk-utils/utils/urand"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

const (
	alice        = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bech32Prefix = "eth"
)

var bond = types.DefaultParams().MinBond

func (suite *SequencerTestSuite) TestMinBond() {
	testCases := []struct {
		name          string
		requiredBond  sdk.Coin
		bond          sdk.Coin
		expectedError error
	}{
		{
			name:          "No bond required",
			requiredBond:  sdk.Coin{},
			bond:          sdk.NewCoin("adym", sdk.NewInt(10000000)),
			expectedError: nil,
		},
		{
			name:          "Valid bond",
			requiredBond:  bond,
			bond:          bond,
			expectedError: nil,
		},
		{
			name:          "Bad denom",
			requiredBond:  bond,
			bond:          sdk.NewCoin("invalid", sdk.NewInt(100)),
			expectedError: types.ErrInvalidCoinDenom,
		},
		{
			name:          "Insufficient bond",
			requiredBond:  bond,
			bond:          sdk.NewCoin(bond.Denom, bond.Amount.Quo(sdk.NewInt(2))),
			expectedError: types.ErrInsufficientBond,
		},
	}

	for _, tc := range testCases {
		seqParams := types.DefaultParams()
		seqParams.MinBond = tc.requiredBond
		suite.App.SequencerKeeper.SetParams(suite.Ctx, seqParams)

		rollappId, pk := suite.CreateDefaultRollapp()

		// fund account
		addr := sdk.AccAddress(pk.Address())
		pkAny, err := codectypes.NewAnyWithValue(pk)
		suite.Require().Nil(err)
		err = bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(tc.bond))
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
		_, err = suite.msgServer.CreateSequencer(suite.Ctx, &sequencerMsg1)
		if tc.expectedError != nil {
			tc := tc
			suite.Require().ErrorAs(err, &tc.expectedError, tc.name)
		} else {
			suite.Require().NoError(err)
			sequencer, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr.String())
			suite.Require().True(found, tc.name)
			if tc.requiredBond.IsNil() {
				suite.Require().True(sequencer.Tokens.IsZero(), tc.name)
			} else {
				suite.Require().Equal(sdk.NewCoins(tc.requiredBond), sequencer.Tokens, tc.name)
			}
		}
	}
}

func (suite *SequencerTestSuite) TestCreateSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	// sequencersExpect is the expected result of query all
	sequencersExpect := []*types.Sequencer{}

	// rollappSequencersExpect is a map from rollappId to a map of sequencer addresses list
	type rollappSequencersExpectKey struct {
		rollappId, sequencerAddress string
	}
	rollappSequencersExpect := make(map[rollappSequencersExpectKey]string)
	rollappExpectedProposers := make(map[string]string)

	const numRollapps = 3
	rollappIDs := make([]string, numRollapps)
	// for 3 rollapps, test 10 sequencers creations
	for j := 0; j < numRollapps; j++ {
		rollapp := rollapptypes.Rollapp{
			RollappId:       urand.RollappID(),
			Owner:           alice,
			Bech32Prefix:    bech32Prefix,
			GenesisChecksum: "1234567890abcdefg",
			Sealed:          true,
			Metadata: &rollapptypes.RollappMetadata{
				Website:     "https://dymension.xyz",
				Description: "Sample description",
				LogoUrl:     "https://dymension.xyz/logo.png",
				Telegram:    "https://t.me/rolly",
				X:           "https://x.dymension.xyz",
			},
		}
		suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

		rollappId := rollapp.GetRollappId()
		rollappIDs[j] = rollappId

		for i := 0; i < 10; i++ {
			pubkey := ed25519.GenPrivKey().PubKey()
			addr := sdk.AccAddress(pubkey.Address())
			err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
			suite.Require().NoError(err)
			pkAny, err := codectypes.NewAnyWithValue(pubkey)
			suite.Require().Nil(err)

			// sequencer is the sequencer to create
			sequencerMsg := types.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         bond,
				RollappId:    rollappId,
				Metadata: types.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}
			// sequencerExpect is the expected result of creating a sequencer
			sequencerExpect := types.Sequencer{
				Address:      sequencerMsg.GetCreator(),
				DymintPubKey: sequencerMsg.GetDymintPubKey(),
				Status:       types.Bonded,
				RollappId:    rollappId,
				Tokens:       sdk.NewCoins(bond),
				Metadata:     sequencerMsg.GetMetadata(),
			}

			// create sequencer
			createResponse, err := suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			suite.Require().Nil(err)
			suite.Require().EqualValues(types.MsgCreateSequencerResponse{}, *createResponse)

			// query the specific sequencer
			queryResponse, err := suite.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{
				SequencerAddress: sequencerMsg.GetCreator(),
			})
			suite.Require().Nil(err)
			suite.equalSequencer(&sequencerExpect, &queryResponse.Sequencer)

			// add the sequencer to the list of get all expected list
			sequencersExpect = append(sequencersExpect, &sequencerExpect)

			if i == 0 {
				rollappExpectedProposers[rollappId] = sequencerExpect.Address
			}

			sequencersRes, totalRes := getAll(suite)
			suite.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			suite.verifyAll(sequencersExpect, sequencersRes)

			// add the sequencer to the list of specific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.Address}] = sequencerExpect.Address
		}
	}

	totalFound := 0
	// check query by rollapp
	for i := 0; i < numRollapps; i++ {
		rollappId := rollappIDs[i]
		queryAllResponse, err := suite.queryClient.SequencersByRollapp(goCtx,
			&types.QueryGetSequencersByRollappRequest{RollappId: rollappId})
		suite.Require().Nil(err)
		// verify that all the addresses of the rollapp are found
		for _, sequencer := range queryAllResponse.Sequencers {
			suite.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencer.Address}],
				sequencer.Address)
		}
		totalFound += len(queryAllResponse.Sequencers)

		// check that the first sequencer created is the active sequencer
		proposer, err := suite.queryClient.GetProposerByRollapp(goCtx,
			&types.QueryGetProposerByRollappRequest{RollappId: rollappId})
		suite.Require().Nil(err)
		suite.Require().EqualValues(proposer.ProposerAddr, rollappExpectedProposers[rollappId])
	}
	suite.Require().EqualValues(totalFound, len(rollappSequencersExpect))
}

func (suite *SequencerTestSuite) TestCreateSequencerAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	rollappId, pk := suite.CreateDefaultRollapp()
	addr := sdk.AccAddress(pk.Address())
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
	suite.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pk)
	suite.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.EqualError(err, types.ErrSequencerExists.Error())

	// unbond the sequencer
	unbondMsg := types.MsgUnbond{Creator: addr.String()}
	_, err = suite.msgServer.Unbond(goCtx, &unbondMsg)
	suite.Require().NoError(err)

	// create the sequencer again, expect to fail anyway
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.EqualError(err, types.ErrSequencerExists.Error())
}

func (suite *SequencerTestSuite) TestCreateSequencerInitialSequencerAsProposer() {
	const alex = "dym1te3lcav5c2jn8tdcrhnyl8aden6lglw266kcdd"

	type sequencer struct {
		creatorName string
		expProposer bool
	}
	testCases := []struct {
		name,
		rollappInitialSeq string
		sequencers []sequencer
		malleate   func(rollappID string)
		expErr     error
	}{
		{
			name:              "Single initial sequencer is the first proposer",
			sequencers:        []sequencer{{creatorName: "alex", expProposer: true}},
			rollappInitialSeq: alex,
		}, {
			name:              "Two sequencers - one is the proposer",
			sequencers:        []sequencer{{creatorName: "alex", expProposer: true}, {creatorName: "bob", expProposer: false}},
			rollappInitialSeq: fmt.Sprintf("%s,%s", alice, alex),
		}, {
			name:              "One sequencer - failed because no initial sequencer",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: alice,
			expErr:            types.ErrNotInitialSequencer,
		}, {
			name:              "Any sequencer can be the first proposer",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: true}, {creatorName: "steve", expProposer: false}},
			rollappInitialSeq: "*",
		}, {
			name:              "success - any sequencer can be the first proposer, rollapp sealed",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: alice,
			malleate: func(rollappID string) {
				r, _ := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappID)
				r.Sealed = true
				suite.App.RollappKeeper.SetRollapp(suite.Ctx, r)
			},
			expErr: nil,
		}, {
			name:              "success - no initial sequencer, rollapp sealed",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: "*",
			malleate: func(rollappID string) {
				r, _ := suite.App.RollappKeeper.GetRollapp(suite.Ctx, rollappID)
				r.Sealed = true
				suite.App.RollappKeeper.SetRollapp(suite.Ctx, r)
			},
			expErr: nil,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		goCtx := sdk.WrapSDKContext(suite.Ctx)
		rollappId := suite.CreateRollappWithInitialSequencer(tc.rollappInitialSeq)

		if tc.malleate != nil {
			tc.malleate(rollappId)
		}

		for _, seq := range tc.sequencers {
			addr, pk := sample.AccFromSecret(seq.creatorName)
			pkAny, _ := codectypes.NewAnyWithValue(pk)

			err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
			suite.Require().NoError(err)

			sequencerMsg := types.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         bond,
				RollappId:    rollappId,
				Metadata: types.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}
			_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			suite.Require().ErrorIs(err, tc.expErr, tc.name)

			if tc.expErr != nil {
				return
			}

			// check that the sequencer is the proposer
			proposer, ok := suite.App.SequencerKeeper.GetProposer(suite.Ctx, rollappId)
			suite.Require().True(ok)
			if seq.expProposer {
				suite.Require().Equal(addr.String(), proposer.Address, tc.name)
			} else {
				suite.Require().NotEqual(addr.String(), proposer.Address, tc.name)
			}
		}
	}
}

func (suite *SequencerTestSuite) TestCreateSequencerUnknownRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	pubkey := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
	suite.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	suite.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    "rollappId",
		Metadata:     types.SequencerMetadata{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.EqualError(err, types.ErrUnknownRollappID.Error())
}

// ---------------------------------------
// verifyAll receives a list of expected results and a map of sequencerAddress->sequencer
// the function verifies that the map contains all the sequencers that are in the list and only them
func (suite *SequencerTestSuite) verifyAll(sequencersExpect []*types.Sequencer, sequencersRes map[string]*types.Sequencer) {
	// check number of items are equal
	suite.Require().EqualValues(len(sequencersExpect), len(sequencersRes))
	for i := 0; i < len(sequencersExpect); i++ {
		sequencerExpect := sequencersExpect[i]
		sequencerRes := sequencersRes[sequencerExpect.GetAddress()]
		suite.equalSequencer(sequencerExpect, sequencerRes)
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
func (suite *SequencerTestSuite) equalSequencer(s1 *types.Sequencer, s2 *types.Sequencer) {
	eq := compareSequencers(s1, s2)
	suite.Require().True(eq, "expected: %v\nfound: %v", *s1, *s2)
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

	if s1.Jailed != s2.Jailed {
		return false
	}
	if s1.Status != s2.Status {
		return false
	}

	if !s1.Tokens.IsEqual(s2.Tokens) {
		return false
	}

	if s1.UnbondRequestHeight != s2.UnbondRequestHeight {
		return false
	}
	if !s1.UnbondTime.Equal(s2.UnbondTime) {
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
