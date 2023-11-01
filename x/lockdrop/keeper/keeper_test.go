package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	queryClient types.QueryClient
}

func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()

	suite.queryClient = types.NewQueryClient(suite.QueryHelper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
