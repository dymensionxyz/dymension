package apptesting

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/stretchr/testify/suite"
)

type KeeperTestHelper struct {
	suite.Suite
	App *app.App
	Ctx sdk.Context
}
