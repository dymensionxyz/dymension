package forward_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
}

func (s *TestSuite) SetupTest() {
	app := apptesting.Setup(s.T())
	ctx := app.BaseApp.NewContext(false)

	s.App = app
	s.Ctx = ctx
}

func TestSequencerKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestTheSetup() {
	s.SetupHyperlane()
}
