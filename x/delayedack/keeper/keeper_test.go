package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"
)

const (
	delayedAckEventType = "delayedack"
)

type DelayedAckTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(DelayedAckTestSuite))
}

func (suite *DelayedAckTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	suite.App = app
	suite.Ctx = ctx
}
