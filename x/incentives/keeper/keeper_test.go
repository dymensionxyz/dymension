package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
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
	app := apptesting.Setup(suite.T())
	ctx := app.NewContext(false)

	suite.App = app
	suite.Ctx = ctx

	suite.querier = keeper.NewQuerier(*suite.App.IncentivesKeeper)
	lockableDurations := suite.App.IncentivesKeeper.GetLockableDurations(suite.Ctx)
	lockableDurations = append(lockableDurations, 2*time.Second)
	suite.App.IncentivesKeeper.SetLockableDurations(suite.Ctx, lockableDurations)

	// Set endorsement mode for all rollapps
	params := suite.App.IncentivesKeeper.GetParams(suite.Ctx)
	params.RollappGaugesMode = types.Params_AllRollapps
	suite.App.IncentivesKeeper.SetParams(suite.Ctx, params)

	_ = suite.App.TxFeesKeeper.SetBaseDenom(suite.Ctx, "adym")
}
