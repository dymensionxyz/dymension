package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/suite"

	keeper "github.com/dymensionxyz/dymension/x/streamer/keeper"
	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	moduleAddress sdk.AccAddress

	querier keeper.Querier
}

// SetupTest sets streamer parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.moduleAddress = authtypes.NewModuleAddress(types.ModuleName)
	suite.querier = keeper.NewQuerier(suite.App.StreamerKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
