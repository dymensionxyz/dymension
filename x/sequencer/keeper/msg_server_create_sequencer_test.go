package keeper_test

import (
	"fmt"
	"reflect"
	"time"

	errorsmod "cosmossdk.io/errors"
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

func (s *SequencerTestSuite) TestMinBond() {
	testCases := []struct {
		name          string
		requiredBond  sdk.Coin
		bond          sdk.Coin
		expectedError error
	}{
		{
			name:          "Valid bond",
			requiredBond:  bond,
			bond:          bond,
			expectedError: nil,
		},
		{
			name:          "Insufficient bond",
			requiredBond:  bond,
			bond:          sdk.NewCoin(bond.Denom, bond.Amount.Quo(sdk.NewInt(2))),
			expectedError: types.ErrInsufficientBond,
		},
		{
			name:          "wrong bond denom",
			requiredBond:  bond,
			bond:          sdk.NewCoin("nonbonddenom", bond.Amount),
			expectedError: types.ErrInvalidDenom,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			seqParams := types.DefaultParams()
			seqParams.MinBond = tc.requiredBond
			s.App.SequencerKeeper.SetParams(s.Ctx, seqParams)

			rollappId, pk := s.CreateDefaultRollapp()

			// fund account
			addr := sdk.AccAddress(pk.Address())
			pkAny, err := codectypes.NewAnyWithValue(pk)
			s.Require().Nil(err)
			err = bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(tc.bond))
			s.Require().Nil(err)

			sequencerMsg1 := types.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         tc.bond,
				RollappId:    rollappId,
				Metadata: types.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}

			// Use a defer and recover to catch potential panics
			var createErr error
			func() {
				defer func() {
					if r := recover(); r != nil {
						createErr = errorsmod.Wrapf(panicErr, "panic: %v", r)
					}
				}()
				_, createErr = s.msgServer.CreateSequencer(s.Ctx, &sequencerMsg1)
			}()

			if tc.expectedError != nil {
				s.Require().ErrorAs(createErr, &tc.expectedError, tc.name)
			} else {
				s.Require().NoError(createErr)
				sequencer, found := s.App.SequencerKeeper.GetSequencer(s.Ctx, addr.String())
				s.Require().True(found, tc.name)
				if tc.requiredBond.IsNil() {
					s.Require().True(sequencer.Tokens.IsZero(), tc.name)
				} else {
					s.Require().Equal(sdk.NewCoins(tc.requiredBond), sequencer.Tokens, tc.name)
				}
			}
		})
	}
}

func (s *SequencerTestSuite) TestCreateSequencer() {
	s.SetupTest()
	goCtx := sdk.WrapSDKContext(s.Ctx)

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
			RollappId: urand.RollappID(),
			Owner:     alice,
			Launched:  true,
			Metadata: &rollapptypes.RollappMetadata{
				Website:     "https://dymension.xyz",
				Description: "Sample description",
				LogoUrl:     "https://dymension.xyz/logo.png",
				Telegram:    "https://t.me/rolly",
				X:           "https://x.dymension.xyz",
			},
			GenesisInfo: rollapptypes.GenesisInfo{
				Bech32Prefix:    bech32Prefix,
				GenesisChecksum: "1234567890abcdefg",
				InitialSupply:   sdk.NewInt(1000),
				NativeDenom: rollapptypes.DenomMetadata{
					Display:  "DEN",
					Base:     "aden",
					Exponent: 18,
				},
			},
		}
		s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

		rollappId := rollapp.GetRollappId()
		rollappIDs[j] = rollappId

		for i := 0; i < 10; i++ {
			pubkey := ed25519.GenPrivKey().PubKey()
			addr := sdk.AccAddress(pubkey.Address())
			err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(bond))
			s.Require().NoError(err)
			pkAny, err := codectypes.NewAnyWithValue(pubkey)
			s.Require().Nil(err)

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
			createResponse, err := s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			s.Require().Nil(err)
			s.Require().EqualValues(types.MsgCreateSequencerResponse{}, *createResponse)

			// query the specific sequencer
			queryResponse, err := s.queryClient.Sequencer(goCtx, &types.QueryGetSequencerRequest{
				SequencerAddress: sequencerMsg.GetCreator(),
			})
			s.Require().Nil(err)
			s.equalSequencer(&sequencerExpect, &queryResponse.Sequencer)

			// add the sequencer to the list of get all expected list
			sequencersExpect = append(sequencersExpect, &sequencerExpect)

			if i == 0 {
				rollappExpectedProposers[rollappId] = sequencerExpect.Address
			}

			sequencersRes, totalRes := getAll(s)
			s.Require().EqualValues(len(sequencersExpect), totalRes)
			// verify that query all contains all the sequencers that were created
			s.verifyAll(sequencersExpect, sequencersRes)

			// add the sequencer to the list of specific rollapp
			rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencerExpect.Address}] = sequencerExpect.Address
		}
	}

	totalFound := 0
	// check query by rollapp
	for i := 0; i < numRollapps; i++ {
		rollappId := rollappIDs[i]
		queryAllResponse, err := s.queryClient.SequencersByRollapp(goCtx,
			&types.QueryGetSequencersByRollappRequest{RollappId: rollappId})
		s.Require().Nil(err)
		// verify that all the addresses of the rollapp are found
		for _, sequencer := range queryAllResponse.Sequencers {
			s.Require().EqualValues(rollappSequencersExpect[rollappSequencersExpectKey{rollappId, sequencer.Address}],
				sequencer.Address)
		}
		totalFound += len(queryAllResponse.Sequencers)

		// check that the first sequencer created is the active sequencer
		proposer, err := s.queryClient.GetProposerByRollapp(goCtx,
			&types.QueryGetProposerByRollappRequest{RollappId: rollappId})
		s.Require().Nil(err)
		s.Require().EqualValues(proposer.ProposerAddr, rollappExpectedProposers[rollappId])
	}
	s.Require().EqualValues(totalFound, len(rollappSequencersExpect))
}

func (s *SequencerTestSuite) TestCreateSequencerAlreadyExists() {
	s.SetupTest()
	goCtx := sdk.WrapSDKContext(s.Ctx)

	rollappId, pk := s.CreateDefaultRollapp()
	addr := sdk.AccAddress(pk.Address())
	err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pk)
	s.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}
	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.Require().Nil(err)

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.EqualError(err, types.ErrSequencerAlreadyExists.Error())

	// unbond the sequencer
	unbondMsg := types.MsgUnbond{Creator: addr.String()}
	_, err = s.msgServer.Unbond(goCtx, &unbondMsg)
	s.Require().NoError(err)

	// create the sequencer again, expect to fail anyway
	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.EqualError(err, types.ErrSequencerAlreadyExists.Error())
}

func (s *SequencerTestSuite) TestCreateSequencerInitialSequencerAsProposer() {
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
			name:              "success - any sequencer can be the first proposer, rollapp launched",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: alice,
			malleate: func(rollappID string) {
				r, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappID)
				r.Launched = true
				s.App.RollappKeeper.SetRollapp(s.Ctx, r)
			},
			expErr: nil,
		}, {
			name:              "success - no initial sequencer, rollapp launched",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: "*",
			malleate: func(rollappID string) {
				r, _ := s.App.RollappKeeper.GetRollapp(s.Ctx, rollappID)
				r.Launched = true
				s.App.RollappKeeper.SetRollapp(s.Ctx, r)
			},
			expErr: nil,
		},
	}

	for _, tc := range testCases {
		s.SetupTest()

		goCtx := sdk.WrapSDKContext(s.Ctx)
		rollappId := s.CreateRollappWithInitialSequencer(tc.rollappInitialSeq)

		if tc.malleate != nil {
			tc.malleate(rollappId)
		}

		for _, seq := range tc.sequencers {
			addr, pk := sample.AccFromSecret(seq.creatorName)
			pkAny, _ := codectypes.NewAnyWithValue(pk)

			err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(bond))
			s.Require().NoError(err)

			sequencerMsg := types.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         bond,
				RollappId:    rollappId,
				Metadata: types.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}
			_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			s.Require().ErrorIs(err, tc.expErr, tc.name)

			if tc.expErr != nil {
				return
			}

			// check that the sequencer is the proposer
			proposer, ok := s.App.SequencerKeeper.GetProposerLegacy(s.Ctx, rollappId)
			s.Require().True(ok)
			if seq.expProposer {
				s.Require().Equal(addr.String(), proposer.Address, tc.name)
			} else {
				s.Require().NotEqual(addr.String(), proposer.Address, tc.name)
			}
		}
	}
}

func (s *SequencerTestSuite) TestCreateSequencerUnknownRollappId() {
	s.SetupTest()
	goCtx := sdk.WrapSDKContext(s.Ctx)

	pubkey := ed25519.GenPrivKey().PubKey()
	addr := sdk.AccAddress(pubkey.Address())
	err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pubkey)
	s.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    "rollappId",
		Metadata:     types.SequencerMetadata{},
	}

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.EqualError(err, types.ErrRollappNotFound.Error())
}

// create sequencer before genesisInfo is set
func (s *SequencerTestSuite) TestCreateSequencerBeforeGenesisInfo() {
	s.SetupTest()
	goCtx := sdk.WrapSDKContext(s.Ctx)
	rollappId, pk := s.CreateDefaultRollapp()

	// mess up the genesisInfo
	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	rollapp.GenesisInfo.Bech32Prefix = ""
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

	addr := sdk.AccAddress(pk.Address())
	err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pk)
	s.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.Require().Error(err)

	// set the genesisInfo
	rollapp.GenesisInfo.Bech32Prefix = "rol"
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

	_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
	s.Require().NoError(err)
}

// create sequencer before prelaunch
func (s *SequencerTestSuite) TestCreateSequencerBeforePrelaunch() {
	s.SetupTest()
	rollappId, pk := s.CreateDefaultRollapp()

	// set prelaunch time to the rollapp
	preLaunchTime := time.Now()
	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	rollapp.PreLaunchTime = &preLaunchTime
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

	addr := sdk.AccAddress(pk.Address())
	err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, sdk.NewCoins(bond))
	s.Require().NoError(err)

	pkAny, err := codectypes.NewAnyWithValue(pk)
	s.Require().Nil(err)
	sequencerMsg := types.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
		RollappId:    rollappId,
		Metadata: types.SequencerMetadata{
			Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
		},
	}

	_, err = s.msgServer.CreateSequencer(sdk.WrapSDKContext(s.Ctx), &sequencerMsg)
	s.Require().Error(err)

	s.Ctx = s.Ctx.WithBlockTime(preLaunchTime.Add(time.Second))
	_, err = s.msgServer.CreateSequencer(sdk.WrapSDKContext(s.Ctx), &sequencerMsg)
	s.Require().NoError(err)
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
	s.Require().True(eq, "expected: %v\nfound: %v", *s1, *s2)
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
