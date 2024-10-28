package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uptr"
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
	s.Require().True(s.k().IsProposer(s.Ctx, seq))
	s.Require().True(equalSequencers(uptr.To(expectedSequencer(&msg)), &seq))
}

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
	s.compareAllSequencersResponse(getOneRes, allRes)
}

// There are several reasons to reject creation
func (s *SequencerTestSuite) TestCreateSequencerRestrictions() {
	ra := s.createRollapp()

	s.Run("not allowed - no rollapp", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(urand.RollappID(), alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("not allowed - already exist", func() {
		pk := randPK()
		s.fundSequencer(pk, bond)
		msg := createSequencerMsg(ra.RollappId, pk)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrAlreadyExists)
	})
	s.Run("not allowed - TODO: awaitingLastProposerBlock", func() {
	})
	s.Run("not allowed - insufficient bond", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Amount = msg.Bond.Amount.Sub(sdk.OneInt())
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrOutOfRange)
	})
	s.Run("not allowed - wrong denom", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Denom = "foo"
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrInvalidArgument)
	})
	s.Run("not allowed - vm", func() {
	})

	s.Run("not allowed - not launched and not initial", func() {
		ra := s.createRollappWithInitialSeqConstraint("")
		launched := s.raK().MustGetRollapp(s.Ctx, ra).Launched
		s.Require().False(launched)

		s.fundSequencer(alice, bond)
		msg := createSequencerMsg(ra, alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)
	})
	s.Run("allowed - launched", func() {
		seq := alice
		ra := s.createRollappWithInitialSeqConstraint("")
		rollapp := s.raK().MustGetRollapp(s.Ctx, ra)
		rollapp.Launched = true
		s.raK().SetRollapp(s.Ctx, rollapp)

		s.fundSequencer(seq, bond)
		msg := createSequencerMsg(ra, seq)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
	})
	s.Run("allowed - initial", func() {
		seq := bob
		ra := s.createRollappWithInitialSeqConstraint(pkAddr(bob))
		launched := s.raK().MustGetRollapp(s.Ctx, ra).Launched
		s.Require().False(launched)

		s.fundSequencer(seq, bond)
		msg := createSequencerMsg(ra, seq)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
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
func (s *SequencerTestSuite) compareAllSequencersResponse(
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
