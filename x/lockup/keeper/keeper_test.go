package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/lockup/keeper"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	querier keeper.Querier
	cleanup func()
}

func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T())
	ctx := app.NewContext(false)

	suite.App = app
	suite.Ctx = ctx

	suite.querier = keeper.NewQuerier(*suite.App.LockupKeeper)
	params, err := suite.App.StakingKeeper.GetParams(suite.Ctx)
	suite.Require().NoError(err)
	unbondingDuration := params.UnbondingTime
	suite.App.IncentivesKeeper.SetLockableDurations(suite.Ctx, []time.Duration{
		time.Hour * 24 * 14,
		time.Hour,
		time.Hour * 3,
		time.Hour * 7,
		unbondingDuration,
	})
}

func (suite *KeeperTestSuite) Cleanup() {
	suite.cleanup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
