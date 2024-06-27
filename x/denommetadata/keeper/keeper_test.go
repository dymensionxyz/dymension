package keeper_test

import (
	"testing"

	errorsmod "cosmossdk.io/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
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
	bankKeeper := suite.App.BankKeeper
	err := keeper.CreateDenomMetadata(suite.Ctx, suite.getDymMetadata())
	suite.Require().NoError(err)

	denom, found := bankKeeper.GetDenomMetaData(suite.Ctx, suite.getDymMetadata().Base)
	suite.Require().EqualValues(found, true)
	suite.Require().EqualValues(denom.Symbol, suite.getDymMetadata().Symbol)
}

func (suite *KeeperTestSuite) TestUpdateDenom() {
	keeper := suite.App.DenomMetadataKeeper
	bankKeeper := suite.App.BankKeeper

	err := keeper.CreateDenomMetadata(suite.Ctx, suite.getDymMetadata())
	suite.Require().NoError(err)

	denom, found := bankKeeper.GetDenomMetaData(suite.Ctx, suite.getDymMetadata().Base)
	suite.Require().EqualValues(found, true)
	suite.Require().EqualValues(denom.DenomUnits[1].Exponent, suite.getDymMetadata().DenomUnits[1].Exponent)

	err = keeper.UpdateDenomMetadata(suite.Ctx, suite.getDymUpdateMetadata())
	suite.Require().NoError(err)

	denom, found = bankKeeper.GetDenomMetaData(suite.Ctx, suite.getDymUpdateMetadata().Base)
	suite.Require().EqualValues(found, true)
	suite.Require().EqualValues(denom.DenomUnits[1].Exponent, suite.getDymUpdateMetadata().DenomUnits[1].Exponent)
}

func (suite *KeeperTestSuite) TestCreateExistingDenom() {
	keeper := suite.App.DenomMetadataKeeper
	err := keeper.CreateDenomMetadata(suite.Ctx, suite.getDymMetadata())
	suite.Require().NoError(err)

	err = keeper.CreateDenomMetadata(suite.Ctx, suite.getDymMetadata())
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrAlreadyExists))
}

func (suite *KeeperTestSuite) TestUpdateMissingDenom() {
	keeper := suite.App.DenomMetadataKeeper

	err := keeper.UpdateDenomMetadata(suite.Ctx, suite.getDymUpdateMetadata())
	suite.Require().True(errorsmod.IsOf(err, gerrc.ErrNotFound))
}

func (suite *KeeperTestSuite) getDymMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Name:        "Dymension Hub token",
		Symbol:      "DYM",
		Description: "Denom metadata for DYM.",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "adym", Exponent: uint32(0), Aliases: []string{}},
			{Denom: "DYM", Exponent: uint32(18), Aliases: []string{}},
		},
		Base:    "adym",
		Display: "DYM",
	}
}

func (suite *KeeperTestSuite) getDymUpdateMetadata() banktypes.Metadata {
	return banktypes.Metadata{
		Name:        "Dymension Hub token",
		Symbol:      "DYM",
		Description: "Denom metadata for DYM.",
		DenomUnits: []*banktypes.DenomUnit{
			{Denom: "adym", Exponent: uint32(0), Aliases: []string{}},
			{Denom: "DYM", Exponent: uint32(9), Aliases: []string{}},
		},
		Base:    "adym",
		Display: "DYM",
	}
}
