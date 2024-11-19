package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
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
	msg := createSequencerMsgOnePubkey(ra.RollappId, alice)
	msg.Bond = bond
	_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
	s.Require().NoError(err)
	seq, err := s.k().RealSequencer(s.Ctx, pkAddr(alice))
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
		msg := createSequencerMsgOnePubkey(ra.RollappId, pk)
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
		msg := createSequencerMsgOnePubkey(urand.RollappID(), alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrNotFound)
	})
	s.Run("not allowed - already exist", func() {
		pk := randomTMPubKey()
		s.fundSequencer(pk, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, pk)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrAlreadyExists)
	})
	s.Run("not allowed - pub key in use", func() {
		pk := randomTMPubKey()
		s.fundSequencer(pk, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, pk)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
		s.fundSequencer(alice, bond)
		msg = createSequencerMsg(ra.RollappId, alice, pk)
		msg.Bond = bond
		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrAlreadyExists)
	})

	s.Run("not allowed - insufficient bond", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Amount = msg.Bond.Amount.Sub(sdk.OneInt())
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrOutOfRange)
	})
	s.Run("not allowed - wrong denom", func() {
		s.fundSequencer(alice, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, alice)
		msg.Bond = bond
		msg.Bond.Denom = "foo"
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrInvalidArgument)
	})
	s.Run("not allowed - vm", func() {
		ra := s.createRollapp()
		ra.VmType = rollapptypes.Rollapp_EVM
		s.raK().SetRollapp(s.Ctx, ra)
		s.fundSequencer(alice, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, alice)
		msg.Bond = bond
		msg.Metadata.EvmRpcs = msg.Metadata.EvmRpcs[:0]
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrInvalidArgument)
	})
	s.Run("not allowed - not launched and not initial", func() {
		ra := s.createRollappWithInitialSeqConstraint("")
		s.Require().False(ra.Launched)

		s.fundSequencer(alice, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, alice)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)
	})

	s.Run("not allowed - pre launch", func() {
		ra := s.createRollappWithInitialSeqConstraint("*")
		s.Require().False(ra.Launched)
		ra.PreLaunchTime = uptr.To(s.Ctx.BlockTime().Add(time.Hour))
		s.raK().SetRollapp(s.Ctx, ra)

		seq := david
		s.fundSequencer(seq, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, seq)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)

		_, err = s.msgServer.CreateSequencer(s.Ctx.WithBlockTime(*ra.PreLaunchTime), &msg)
		s.Require().NoError(err)
	})

	s.Run("allowed - launched", func() {
		seq := alice
		ra := s.createRollappWithInitialSeqConstraint("")
		ra.Launched = true
		s.raK().SetRollapp(s.Ctx, ra)

		s.fundSequencer(seq, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, seq)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
	})
	s.Run("allowed - pre launch and is initial", func() {
		seq := bob
		ra := s.createRollappWithInitialSeqConstraint(pkAddr(bob))
		s.Require().False(ra.Launched)

		s.fundSequencer(seq, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, seq)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
	})
	s.Run("not allowed - genesis info broken", func() {
		seq := charlie
		ra := s.createRollapp()
		ra.GenesisInfo.Bech32Prefix = ""
		s.raK().SetRollapp(s.Ctx, ra)

		s.fundSequencer(seq, bond)
		msg := createSequencerMsgOnePubkey(ra.RollappId, seq)
		msg.Bond = bond
		_, err := s.msgServer.CreateSequencer(s.Ctx, &msg)
		utest.IsErr(s.Require(), err, gerrc.ErrFailedPrecondition)

		ra.GenesisInfo.Bech32Prefix = "eth"
		s.raK().SetRollapp(s.Ctx, ra)

		_, err = s.msgServer.CreateSequencer(s.Ctx, &msg)
		s.Require().NoError(err)
	})
}

func expectedSequencer(m *types.MsgCreateSequencer) types.Sequencer {
	return types.Sequencer{
		Address:          m.Creator,
		DymintPubKey:     m.DymintPubKey,
		RollappId:        m.RollappId,
		Metadata:         m.Metadata,
		Status:           types.Bonded,
		OptedIn:          true,
		Tokens:           sdk.NewCoins(m.Bond),
		NoticePeriodTime: time.Time{},
		RewardAddr:       m.Creator,
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
