package keeper_test

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

const (
	alice         = "cosmos1jmjfq0tplp9tmx4v9uemw72y4d2wa5nr3xn9d3"
	bob           = "cosmos1xyxs3skf3f4jfqeuv89yyaqvjc6lffavxqhc8g"
	carol         = "cosmos1e0w5t53nrq7p66fye6c8p0ynyhf6y24l4yuxd7"
	balAlice      = 50000000
	balBob        = 20000000
	balCarol      = 10000000
	foreignToken  = "foreignToken"
	balTokenAlice = 5
	balTokenBob   = 2
	balTokenCarol = 1
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
			NoticePeriod:  10,
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
			Description:  types.Description{},
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

	// max sequencers per rollapp
	maxSequencers := 10

	// sequencersExpect is the expected result of query all
	sequencersExpect := []*types.Sequencer{}

	// rollappSequencersExpect is a map from rollappId to a map of sequencer addresses list
	type rollappSequencersExpectKey struct {
		rollappId, sequencerAddress string
	}
	rollappSequencersExpect := make(map[rollappSequencersExpectKey]string)
	rollappExpectedProposers := make(map[string]string)

	// for 3 rollapps, test 10 sequencers creations
	for j := 0; j < 3; j++ {
		rollapp := rollapptypes.Rollapp{
			RollappId:             fmt.Sprintf("%s%d", "rollapp", j),
			Creator:               alice,
			Version:               0,
			MaxSequencers:         uint64(maxSequencers),
			PermissionedAddresses: []string{},
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
				Description:  types.Description{},
			}
			// sequencerExpect is the expected result of creating a sequencer
			sequencerExpect := types.Sequencer{
				SequencerAddress: sequencerMsg.GetCreator(),
				DymintPubKey:     sequencerMsg.GetDymintPubKey(),
				Status:           types.Bonded,
				RollappId:        rollappId,
				Tokens:           sdk.NewCoins(bond),
				Description:      sequencerMsg.GetDescription(),
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

			if i == 0 {
				rollappExpectedProposers[rollappId] = sequencerExpect.SequencerAddress
			}

			sequencersRes, totalRes := getAll(suite)
			suite.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			verifyAll(suite, sequencersExpect, sequencersRes)

			// add the sequencer to the list of specific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.SequencerAddress}] = sequencerExpect.SequencerAddress
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
			suite.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencer.SequencerAddress}],
				sequencer.SequencerAddress)
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

	rollappId := suite.CreateDefaultRollapp()
	_ = suite.CreateDefaultSequencer(suite.Ctx, rollappId) // proposer

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
		Description:  types.Description{},
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
		Description:  types.Description{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.EqualError(err, types.ErrUnknownRollappID.Error())
}

func (suite *SequencerTestSuite) TestCreatePermissionedSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	sequencerAddress := addr.String()
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
	suite.Require().NoError(err)

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         1,
		PermissionedAddresses: []string{sequencerAddress},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)
	rollappId := rollapp.GetRollappId()

	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	suite.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      sequencerAddress,
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Description:  types.Description{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.Require().Nil(err)

	// query the specific sequencer
	queryResponse, err := suite.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{
		SequencerAddress: sequencerMsg.GetCreator(),
	})
	suite.Require().Nil(err)

	// sequencerExpect is the expected result of creating a sequencer
	sequencerExpect := types.Sequencer{
		SequencerAddress: sequencerMsg.GetCreator(),
		DymintPubKey:     sequencerMsg.GetDymintPubKey(),
		Status:           types.Bonded,
		RollappId:        rollappId,
		Description:      sequencerMsg.GetDescription(),
		Tokens:           sdk.NewCoins(bond),
	}
	equalSequencer(suite, &sequencerExpect, &queryResponse.Sequencer)
}

func (suite *SequencerTestSuite) TestCreateSequencerNotPermissioned() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         1,
		PermissionedAddresses: []string{sample.AccAddress()},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	// TODO: use common func (CreateSequencerWithBond)
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
		Description:  types.Description{},
	}

	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.EqualError(err, types.ErrSequencerNotPermissioned.Error())
}

func (suite *SequencerTestSuite) TestMaxSequencersZero() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	maxSequencers := 0

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         uint64(maxSequencers),
		PermissionedAddresses: []string{},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)
	rollappId := rollapp.GetRollappId()

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
	suite.Require().Nil(err)
	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	suite.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Description:  types.Description{},
	}
	_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	suite.Require().Nil(err)
}

func (suite *SequencerTestSuite) TestMaxSequencersLimit() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)
	maxSequencers := 3

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         uint64(maxSequencers),
		PermissionedAddresses: []string{},
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	// create MaxSequencers
	for i := 0; i < maxSequencers; i++ {
		pubkey := secp256k1.GenPrivKey().PubKey()
		addr := sdk.AccAddress(pubkey.Address())
		err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
		suite.Require().Nil(err)

		pkAny, err := codectypes.NewAnyWithValue(pubkey)
		suite.Require().Nil(err)
		sequencerMsg := types.MsgCreateSequencer{
			Creator:      addr.String(),
			DymintPubKey: pkAny,
			Bond:         bond,
			RollappId:    rollappId,
			Description:  types.Description{},
		}
		_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
		suite.Require().Nil(err)
	}

	// add more to be failed
	for i := 0; i < 2; i++ {
		pubkey := secp256k1.GenPrivKey().PubKey()
		addr := sdk.AccAddress(pubkey.Address())
		err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
		suite.Require().Nil(err)
		pkAny, err := codectypes.NewAnyWithValue(pubkey)
		suite.Require().Nil(err)
		sequencerMsg := types.MsgCreateSequencer{
			Creator:      addr.String(),
			DymintPubKey: pkAny,
			Bond:         bond,
			RollappId:    rollappId,
			Description:  types.Description{},
		}
		_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
		suite.EqualError(err, types.ErrMaxSequencersLimit.Error())
	}
}

func (suite *SequencerTestSuite) TestMaxSequencersNotSet() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.Ctx)

	rollapp := rollapptypes.Rollapp{
		RollappId: "rollapp1",
		Creator:   alice,
		Version:   0,
	}
	suite.App.RollappKeeper.SetRollapp(suite.Ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	// create sequencers
	for i := 0; i < 10; i++ {
		pubkey := secp256k1.GenPrivKey().PubKey()
		addr := sdk.AccAddress(pubkey.Address())
		err := bankutil.FundAccount(suite.App.BankKeeper, suite.Ctx, addr, sdk.NewCoins(bond))
		suite.Require().Nil(err)

		pkAny, err := codectypes.NewAnyWithValue(pubkey)
		suite.Require().Nil(err)
		sequencerMsg := types.MsgCreateSequencer{
			Creator:      addr.String(),
			DymintPubKey: pkAny,
			Bond:         bond,
			RollappId:    rollappId,
			Description:  types.Description{},
		}
		_, err = suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
		suite.Require().Nil(err)
	}
}

// ---------------------------------------
// verifyAll receives a list of expected results and a map of sequencerAddress->sequencer
// the function verifies that the map contains all the sequencers that are in the list and only them
func verifyAll(suite *SequencerTestSuite, sequencersExpect []*types.Sequencer, sequencersRes map[string]*types.Sequencer) {
	// check number of items are equal
	suite.Require().EqualValues(len(sequencersExpect), len(sequencersRes))
	for i := 0; i < len(sequencersExpect); i++ {
		sequencerExpect := sequencersExpect[i]
		sequencerRes := sequencersRes[sequencerExpect.GetSequencerAddress()]
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
			sequencersRes[sequencerRes.GetSequencerAddress()] = &sequencerRes
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
	if s1.SequencerAddress != s2.SequencerAddress {
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
	return true
}
