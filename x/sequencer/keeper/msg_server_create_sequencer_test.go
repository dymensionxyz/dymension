package keeper_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

const (
	alice        = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bech32Prefix = "eth"
)

var bond = types.DefaultParams().MinBond

func (suite *SequencerTestSuite) TestMinBond() {
	suite.SetupTest()
	rollappId := suite.CreateDefaultRollapp()

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
		seqParams := types.Params{
			MinBond:       tc.requiredBond,
			UnbondingTime: 100,
		}
		suite.App.SequencerKeeper.SetParams(suite.Ctx, seqParams)

		pubkey1 := secp256k1.GenPrivKey().PubKey()
		addr1 := sdk.AccAddress(pubkey1.Address())
		pkAny1, err := codectypes.NewAnyWithValue(pubkey1)
		suite.Require().Nil(err)

		// fund account
		err = bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr1, sdk.NewCoins(tc.bond))
		suite.Require().Nil(err)

		sequencerMsg1 := types.MsgCreateSequencer{
			Creator:      addr1.String(),
			DymintPubKey: pkAny1,
			Bond:         bond,
			RollappId:    rollappId,
			Metadata:     types.SequencerMetadata{},
		}
		_, err = suite.msgServer.CreateSequencer(suite.Ctx, &sequencerMsg1)
		if tc.expectedError != nil {
			tc := tc
			suite.Require().ErrorAs(err, &tc.expectedError, tc.name)
		} else {
			suite.Require().NoError(err)
			sequencer, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1.String())
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

	// for 3 rollapps, test 10 sequencers creations
	for j := 0; j < 3; j++ {
		rollapp := rollapptypes.Rollapp{
			RollappId:    fmt.Sprintf("%s%d", "rollapp", j),
			Creator:      alice,
			Bech32Prefix: bech32Prefix,
		}
		suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

		rollappId := rollapp.GetRollappId()

		for i := 0; i < 10; i++ {
			pubkey := secp256k1.GenPrivKey().PubKey()
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
				Metadata:     types.SequencerMetadata{},
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
			if i == 0 {
				sequencerExpect.Status = types.Bonded
				sequencerExpect.Proposer = true
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
			equalSequencer(suite, &sequencerExpect, &queryResponse.Sequencer)

			// add the sequencer to the list of get all expected list
			sequencersExpect = append(sequencersExpect, &sequencerExpect)

			sequencersRes, totalRes := getAll(suite)
			suite.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			verifyAll(suite, sequencersExpect, sequencersRes)

			// add the sequencer to the list of spesific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.Address}] = sequencerExpect.Address
		}
	}

	totalFound := 0
	// check query by rollapp
	for j := 0; j < 3; j++ {
		rollappId := fmt.Sprintf("%s%d", "rollapp", j)
		queryAllResponse, err := suite.queryClient.SequencersByRollapp(goCtx,
			&types.QueryGetSequencersByRollappRequest{RollappId: rollappId})
		suite.Require().Nil(err)
		// verify that all the addresses of the rollapp are found
		for _, sequencer := range queryAllResponse.Sequencers {
			suite.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencer.Address}],
				sequencer.Address)
		}
		totalFound += len(queryAllResponse.Sequencers)
	}
	suite.Require().EqualValues(totalFound, len(rollappSequencersExpect))
}

func (suite *SequencerTestSuite) TestCreateSequencerAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	rollappId := suite.CreateDefaultRollapp()

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
	suite.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	suite.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata:     types.SequencerMetadata{},
	}
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.Require().Nil(err)

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.EqualError(err, types.ErrSequencerExists.Error())
}

func (suite *SequencerTestSuite) TestCreateSequencerUnknownRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	pubkey := secp256k1.GenPrivKey().PubKey()
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

func (suite *SequencerTestSuite) TestUpdateStateSecondSeqErrNotActiveSequencer() {
	suite.SetupTest()

	rollappId := suite.CreateDefaultRollapp()

	// create first sequencer
	addr1 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	// create second sequencer
	addr2 := suite.CreateDefaultSequencer(suite.Ctx, rollappId)

	// check scheduler operating status
	scheduler, found := suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr1)
	suite.Require().True(found)
	suite.EqualValues(scheduler.Status, types.Bonded)
	suite.True(scheduler.Proposer)

	// check scheduler operating status
	scheduler, found = suite.App.SequencerKeeper.GetSequencer(suite.Ctx, addr2)
	suite.Require().True(found)
	suite.EqualValues(scheduler.Status, types.Bonded)
	suite.False(scheduler.Proposer)
}

// ---------------------------------------
// verifyAll receives a list of expected results and a map of sequencerAddress->sequencer
// the function verifies that the map contains all the sequencers that are in the list and only them
func verifyAll(suite *SequencerTestSuite, sequencersExpect []*types.Sequencer, sequencersRes map[string]*types.Sequencer) {
	// check number of items are equal
	suite.Require().EqualValues(len(sequencersExpect), len(sequencersRes))
	for i := 0; i < len(sequencersExpect); i++ {
		sequencerExpect := sequencersExpect[i]
		sequencerRes := sequencersRes[sequencerExpect.GetAddress()]
		equalSequencer(suite, sequencerExpect, sequencerRes)
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

// equalSequencer receives two sequencers and compares them. If there they not equal, fails the test
func equalSequencer(suite *SequencerTestSuite, s1 *types.Sequencer, s2 *types.Sequencer) {
	eq := CompareSequencers(s1, s2)
	suite.Require().True(eq, "expected: %v\nfound: %v", *s1, *s2)
}

func CompareSequencers(s1, s2 *types.Sequencer) bool {
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

	if s1.UnbondingHeight != s2.UnbondingHeight {
		return false
	}
	if !s1.UnbondTime.Equal(s2.UnbondTime) {
		return false
	}
	return true
}
