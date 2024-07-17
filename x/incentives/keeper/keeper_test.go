package keeper_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	querier keeper.Querier
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// SetupTest sets incentives parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	suite.App = app
	suite.Ctx = ctx

	suite.querier = keeper.NewQuerier(*suite.App.IncentivesKeeper)
	lockableDurations := suite.App.IncentivesKeeper.GetLockableDurations(suite.Ctx)
	lockableDurations = append(lockableDurations, 2*time.Second)
	suite.App.IncentivesKeeper.SetLockableDurations(suite.Ctx, lockableDurations)
}
