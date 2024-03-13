package keeper_test

import (
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	suite.App = app
	suite.Ctx = ctx
}

func (suite *KeeperTestSuite) TestCreateDenom() {
	keeper := suite.App.DenomMetadataKeeper

	err := keeper.CreateDenomMetadata(suite.Ctx, suite.getTestMetadata())
	suite.Require().NoError(err)

}

func (suite *KeeperTestSuite) TestUpdateDenom() {
	keeper := suite.App.DenomMetadataKeeper

	err := keeper.UpdateDenomMetadata(suite.Ctx, suite.getTestMetadata())
	suite.Require().NoError(err)
}

func (suite *KeeperTestSuite) getTestMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "uatom", Exponent: uint32(0), Aliases: []string{"microatom"}},
			{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
			{Denom: "atom", Exponent: uint32(6), Aliases: nil},
		},
		Base:    "uatom",
		Display: "atom",
	}
}
