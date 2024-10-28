package keeper_test

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
)

// On success, we should get back an object with all the right info
func (s *SequencerTestSuite) TestCreateSequencerBasic() {
	ra := s.createRollapp()
	s.fundSequencer(alice, bond)
	msg := createSequencerMsg(ra.RollappId, alice)
	msg.Bond = bond
	_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
	s.Require().NoError(err)
	seq, err := s.k().GetRealSequencer(s.Ctx, pkAddr(alice))
	s.Require().NoError(err)
	s.Require().Equal(seq.Address, pkAddr(alice))
	s.Require().True(bond.Equal(seq.TokensCoin()))
	s.Require().Equal(s.moduleBalance(), bond)
}

// On success, we should get back an object with all the right info
func (s *SequencerTestSuite) TestCreateSequencerSeveral() {
	ra := s.createRollapp()

	for _, pk := range pks {
		s.fundSequencer(pk, bond)
		msg := createSequencerMsg(ra.RollappId, pk)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
	}

	getOneRes := make([]*types.Sequencer, 0)

	for _, pk := range pks {
		res, err := s.queryClient.Sequencer(
			sdk.WrapSDKContext(s.Ctx),
			&types.QueryGetSequencerRequest{
				SequencerAddress: pkAddr(pk),
			},
		)
		s.Require().NoError(err)
		getOneRes = append(getOneRes, &res.Sequencer)
	}

	allRes, cnt := getAllSequencersMap(s)
	s.Require().Equal(cnt, len(allRes))
	s.verifyAll(getOneRes, allRes)
}

// On success, we should get back an object with all the right info
func (s *SequencerTestSuite) TestCreateSequencerProposer() {
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
			rollappInitialSeq: fmt.Sprintf("%s,%s", aliceAddr, alex),
		}, {
			name:              "One sequencer - failed because not initial sequencer",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: aliceAddr,
			expErr:            types.ErrNotInitialSequencer,
		}, {
			name:              "Any sequencer can be the first proposer",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: true}, {creatorName: "steve", expProposer: false}},
			rollappInitialSeq: "*",
		}, {
			name:              "success - any sequencer can be the first proposer, rollapp launched",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: aliceAddr,
			malleate: func(rollappID string) {
				r, _ := s.raK().GetRollapp(s.Ctx, rollappID)
				r.Launched = true
				s.raK().SetRollapp(s.Ctx, r)
			},
			expErr: nil,
		}, {
			name:              "success - no initial sequencer, rollapp launched",
			sequencers:        []sequencer{{creatorName: "bob", expProposer: false}},
			rollappInitialSeq: "*",
			malleate: func(rollappID string) {
				r, _ := s.raK().GetRollapp(s.Ctx, rollappID)
				r.Launched = true
				s.raK().SetRollapp(s.Ctx, r)
			},
			expErr: nil,
		},
	}

	for _, tc := range testCases {

		goCtx := types.WrapSDKContext(s.Ctx)
		rollappId := s.createRollapp(tc.rollappInitialSeq)

		if tc.malleate != nil {
			tc.malleate(rollappId)
		}

		for _, seq := range tc.sequencers {
			addr, pk := sample.AccFromSecret(seq.creatorName)
			pkAny, _ := types3.NewAnyWithValue(pk)

			err := testutil.FundAccount(s.App.BankKeeper, s.Ctx, addr, types.NewCoins(bond))
			s.Require().NoError(err)

			sequencerMsg := types2.MsgCreateSequencer{
				Creator:      addr.String(),
				DymintPubKey: pkAny,
				Bond:         bond,
				RollappId:    rollappId,
				Metadata: types2.SequencerMetadata{
					Rpcs: []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
				},
			}
			_, err = s.msgServer.CreateSequencer(goCtx, &sequencerMsg)
			s.Require().ErrorIs(err, tc.expErr, tc.name)

			if tc.expErr != nil {
				return
			}

			// check that the sequencer is the proposer
			proposer := s.k().GetProposer(s.Ctx, rollappId)
			if seq.expProposer {
				s.Require().Equal(addr.String(), proposer.Address, tc.name)
			} else {
				s.Require().NotEqual(addr.String(), proposer.Address, tc.name)
			}
		}
	}
}

// There are several reasons to reject creation
func (s *SequencerTestSuite) TestCreateSequencerRestrictions() {
	ra := s.createRollapp()

	s.Run("no rollapp", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(urand.RollappID(), alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("already exist", func() {
		pk := randPK()
		s.fundSequencer(pk, bond)
		msg := createSequencerMsg(ra.RollappId, pk)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrAlreadyExists)
	})
	s.Run("TODO: awaitingLastProposerBlock", func() {
	})
	s.Run("insufficient bond", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Amount = msg.Bond.Amount.Sub(sdk.OneInt())
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrOutOfRange)
	})
	s.Run("wrong denom", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Denom = "foo"
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrInvalidArgument)
	})
	s.Run("TODO: vm", func() {
	})
	s.Run("TODO: launched", func() {
	})
}

func expectedSequencer(m *types.MsgCreateSequencer) types.Sequencer {
	return types.Sequencer{
		Address:          m.Creator,
		DymintPubKey:     m.DymintPubKey,
		RollappId:        m.RollappId,
		Metadata:         m.Metadata,
		Proposer:         false,
		Status:           types.Bonded,
		OptedIn:          true,
		Tokens:           sdk.NewCoins(m.Bond),
		NoticePeriodTime: time.Time{},
	}
}

// returns <addr -> sequencer, total>
func getAllSequencersMap(suite *SequencerTestSuite) (map[string]*types.Sequencer, int) {
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

// ---------------------------------------
// verifyAll receives a list of expected results and a map of sequencerAddress->sequencer
// the function verifies that the map contains all the sequencers that are in the list and only them
func (s *SequencerTestSuite) verifyAll(
	expected []*types.Sequencer,
	got map[string]*types.Sequencer,
) {
	// check number of items are equal
	s.Require().EqualValues(len(expected), len(got))
	for i := 0; i < len(expected); i++ {
		exp := expected[i]
		res := got[exp.GetAddress()]
		s.equalSequencers(exp, res)
	}
}
