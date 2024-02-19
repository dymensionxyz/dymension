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

var (
	bond = types.DefaultParams().MinBond
)

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
		suite.app.SequencerKeeper.SetParams(suite.ctx, seqParams)

		pubkey1 := secp256k1.GenPrivKey().PubKey()
		addr1 := sdk.AccAddress(pubkey1.Address())
		pkAny1, err := codectypes.NewAnyWithValue(pubkey1)
		suite.Require().Nil(err)

		//fund account
		err = bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr1, sdk.NewCoins(tc.bond))
		suite.Require().Nil(err)

		sequencerMsg1 := types.MsgCreateSequencer{
			Creator:      addr1.String(),
			DymintPubKey: pkAny1,
			Bond:         bond,
			RollappId:    rollappId,
			Description:  types.Description{},
		}
		_, err = suite.msgServer.CreateSequencer(suite.ctx, &sequencerMsg1)
		if tc.expectedError != nil {
			suite.Require().ErrorAs(err, &tc.expectedError, tc.name)
		} else {
			suite.Require().NoError(err)
			sequencer, found := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1.String())
			suite.Require().True(found, tc.name)
			if tc.requiredBond.IsNil() {
				suite.Require().True(sequencer.Tokens.IsZero(), tc.name)
			} else {
				suite.Require().Equal(tc.requiredBond, sequencer.Tokens, tc.name)
			}
		}
	}
}

func (suite *SequencerTestSuite) TestCreateSequencer() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	// max sequencers per rollapp
	maxSequencers := 10

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
			RollappId:             fmt.Sprintf("%s%d", "rollapp", j),
			Creator:               alice,
			Version:               0,
			MaxSequencers:         uint64(maxSequencers),
			PermissionedAddresses: []string{},
		}
		suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

		rollappId := rollapp.GetRollappId()

		for i := 0; i < 10; i++ {
			pubkey := secp256k1.GenPrivKey().PubKey()
			addr := sdk.AccAddress(pubkey.Address())
			bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))
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
				Tokens:           bond,
				Description:      sequencerMsg.GetDescription(),
			}
			if i == 0 {
				sequencerExpect.Status = types.Proposer
			}
			// create sequencer
			createResponse, err := suite.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			suite.Require().Nil(err)
			suite.Require().EqualValues(types.MsgCreateSequencerResponse{}, *createResponse)

			// query the spesific sequencer
			queryResponse, err := suite.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{
				SequencerAddress: sequencerMsg.GetCreator(),
			})
			suite.Require().Nil(err)
			equalSequencer(suite, &sequencerExpect, &queryResponse.SequencerInfo.Sequencer)

			// add the sequencer to the list of get all expected list
			sequencersExpect = append(sequencersExpect, &sequencerExpect)

			sequencersRes, totalRes := getAll(suite)
			suite.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			verifyAll(suite, sequencersExpect, sequencersRes)

			// add the sequencer to the list of spesific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.SequencerAddress}] =
				sequencerExpect.SequencerAddress
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
		for _, sequencerInfo := range queryAllResponse.SequencerInfoList {
			sequencerResAddresses := sequencerInfo.Sequencer.SequencerAddress
			suite.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerResAddresses}],
				sequencerResAddresses)
		}
		totalFound += len(queryAllResponse.SequencerInfoList)
	}
	suite.Require().EqualValues(totalFound, len(rollappSequencersExpect))
}

// TODO: test with different sequencer status
func (suite *SequencerTestSuite) TestCreateSequencerAlreadyExists() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	rollappId := suite.CreateDefaultRollapp()

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))

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
}

func (suite *SequencerTestSuite) TestCreateSequencerUnknownRollappId() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))

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
	goCtx := sdk.WrapSDKContext(suite.ctx)

	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	sequencerAddress := addr.String()
	bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         1,
		PermissionedAddresses: []string{sequencerAddress},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)
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
		Status:           types.Proposer,
		RollappId:        rollappId,
		Description:      sequencerMsg.GetDescription(),
		Tokens:           bond,
	}
	equalSequencer(suite, &sequencerExpect, &queryResponse.SequencerInfo.Sequencer)
}

func (suite *SequencerTestSuite) TestCreateSequencerNotPermissioned() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         1,
		PermissionedAddresses: []string{sample.AccAddress()},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	//TODO: cahnge with common func
	pubkey := secp256k1.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))

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

func (suite *SequencerTestSuite) TestMaxSequencersLimit() {
	suite.SetupTest()
	goCtx := sdk.WrapSDKContext(suite.ctx)
	maxSequencers := 3

	rollapp := rollapptypes.Rollapp{
		RollappId:             "rollapp1",
		Creator:               alice,
		Version:               0,
		MaxSequencers:         uint64(maxSequencers),
		PermissionedAddresses: []string{},
	}
	suite.app.RollappKeeper.SetRollapp(suite.ctx, rollapp)

	rollappId := rollapp.GetRollappId()

	// create MaxSequencers
	for i := 0; i < maxSequencers; i++ {
		pubkey := secp256k1.GenPrivKey().PubKey()
		addr := sdk.AccAddress(pubkey.Address())
		err := bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))
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
		err := bankutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.NewCoins(bond))
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

func (suite *SequencerTestSuite) TestUpdateStateSecondSeqErrNotActiveSequencer() {
	suite.SetupTest()

	rollappId := suite.CreateDefaultRollapp()

	// create first sequencer
	addr1 := suite.CreateDefaultSequencer(suite.ctx, rollappId)

	// create second sequencer
	addr2 := suite.CreateDefaultSequencer(suite.ctx, rollappId)

	// check scheduler operating status
	scheduler, found := suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr1)
	suite.Require().True(found)
	suite.EqualValues(scheduler.Status, types.Proposer)

	// check scheduler operating status
	scheduler, found = suite.app.SequencerKeeper.GetSequencer(suite.ctx, addr2)
	suite.Require().True(found)
	suite.EqualValues(scheduler.Status, types.Bonded)
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

// getAll quires for all exsisting sequencers and returns a map of sequencerId->sequencer
func getAll(suite *SequencerTestSuite) (map[string]*types.Sequencer, int) {
	goCtx := sdk.WrapSDKContext(suite.ctx)
	totalChecked := 0
	totalRes := 0
	nextKey := []byte{}
	sequencersRes := make(map[string]*types.Sequencer)
	for {
		queryAllResponse, err := suite.queryClient.SequencerAll(goCtx,
			&types.QueryAllSequencerRequest{
				Pagination: &query.PageRequest{
					Key:        nextKey,
					Offset:     0,
					Limit:      0,
					CountTotal: true,
					Reverse:    false,
				}})
		suite.Require().Nil(err)

		if totalRes == 0 {
			totalRes = int(queryAllResponse.GetPagination().GetTotal())
		}

		for i := 0; i < len(queryAllResponse.SequencerInfoList); i++ {
			sequencerRes := queryAllResponse.SequencerInfoList[i].Sequencer
			sequencersRes[sequencerRes.GetSequencerAddress()] = &sequencerRes
		}
		totalChecked += len(queryAllResponse.SequencerInfoList)
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

	if s1.Tokens.IsNil() || s2.Tokens.IsNil() {
		if !s1.Tokens.IsNil() || !s2.Tokens.IsNil() {
			return false
		}
	} else if !s1.Tokens.Equal(s2.Tokens) {
		return false
	}

	if s1.UnbondingHeight != s2.UnbondingHeight {
		return false
	}
	if !s1.UnbondingTime.Equal(s2.UnbondingTime) {
		return false
	}
	return true
}
