package keeper

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

type TestSuite struct {
	apptesting.KeeperTestHelper
	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func TestUnbondCondition(t *testing.T) {
	// What should the test be like?
	// first there are
}
