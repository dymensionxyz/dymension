package hooks_test

import (
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/stretchr/testify/suite"
)

type HooksTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestHooksTestSuite(t *testing.T) {
	suite.Run(t, new(HooksTestSuite))
}

func (suite *HooksTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	suite.App = app
	suite.Ctx = ctx
}
