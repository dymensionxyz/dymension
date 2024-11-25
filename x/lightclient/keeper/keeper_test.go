package keeper_test

import (
	"testing"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/sdk-utils/utils/utest"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *TestSuite) SetupTest() {
	app := apptesting.Setup(s.T())
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	s.App = app
	s.Ctx = ctx
}

func (s *TestSuite) k() *keeper.Keeper {
	return &s.App.LightClientKeeper
}

func TestSequencerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

// Basic flow should prevent unbonding at appropriate times, and
// handle pruning.
func (s *TestSuite) TestUnbondConditionFlow() {
	seq := keepertest.Alice

	client := keepertest.CanonClientID

	s.k().SetCanonicalClient(s.Ctx, seq.RollappId, client)

	// allowed!
	err := s.k().CanUnbond(s.Ctx, seq)
	s.Require().NoError(err)

	// add some unverified headers
	for h := range 10 {
		err := s.k().SaveSigner(s.Ctx, seq.Address, client, uint64(h))
		s.Require().NoError(err)
	}

	// not allowed!
	err = s.k().CanUnbond(s.Ctx, seq)
	utest.IsErr(s.Require(), err, sequencertypes.ErrUnbondNotAllowed)

	// we prune some, but still not allowed
	err = s.k().PruneSignersAbove(s.Ctx, client, 6)
	s.Require().NoError(err)

	err = s.k().CanUnbond(s.Ctx, seq)
	utest.IsErr(s.Require(), err, sequencertypes.ErrUnbondNotAllowed)

	// the rest are verified
	for h := range 7 {
		err := s.k().RemoveSigner(s.Ctx, seq.Address, client, uint64(h))
		s.Require().NoError(err)
	}

	// allowed!
	err = s.k().CanUnbond(s.Ctx, seq)
	s.Require().NoError(err)
}

// Basic flow should prevent unbonding at appropriate times, and
// handle pruning.
func (s *TestSuite) TestPruneBelow() {
	seq := keepertest.Alice

	client := keepertest.CanonClientID

	s.k().SetCanonicalClient(s.Ctx, seq.RollappId, client)

	// allowed!
	err := s.k().CanUnbond(s.Ctx, seq)
	s.Require().NoError(err)

	// add some unverified headers
	for h := range 10 {
		err := s.k().SaveSigner(s.Ctx, seq.Address, client, uint64(h))
		s.Require().NoError(err)
	}

	// not allowed!
	err = s.k().CanUnbond(s.Ctx, seq)
	utest.IsErr(s.Require(), err, sequencertypes.ErrUnbondNotAllowed)

	// we prune some, but still not allowed
	err = s.k().PruneSignersBelow(s.Ctx, client, 6)
	s.Require().NoError(err)

	err = s.k().CanUnbond(s.Ctx, seq)
	utest.IsErr(s.Require(), err, sequencertypes.ErrUnbondNotAllowed)

	// the rest are verified
	for h := 6; h < 10; h++ {
		err := s.k().RemoveSigner(s.Ctx, seq.Address, client, uint64(h))
		s.Require().NoError(err)
	}

	// allowed!
	err = s.k().CanUnbond(s.Ctx, seq)
	s.Require().NoError(err)
}
