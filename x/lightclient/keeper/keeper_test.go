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
	app := apptesting.Setup(s.T(), false)
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

func (s *TestSuite) TestUnbondConditionBasicFlow() {
	// What should the test be like?
	// first there are

	seq := keepertest.Alice

	client := keepertest.CanonClientID

	s.k().SetCanonicalClient(s.Ctx, seq.RollappId, client)

	err := s.k().CanUnbond(s.Ctx, seq)
	s.Require().NoError(err)

	for h := range 10 {
		err := s.k().SaveSigner(s.Ctx, seq.Address, client, uint64(h))
		s.Require().NoError(err)
	}

	err = s.k().CanUnbond(s.Ctx, seq)
	utest.IsErr(s.Require(), err, sequencertypes.ErrUnbondNotAllowed)

	for h := range 10 {
		err := s.k().RemoveSigner(s.Ctx, seq.Address, client, uint64(h))
		s.Require().NoError(err)
	}

	err = s.k().CanUnbond(s.Ctx, seq)
	s.Require().NoError(err)
}
