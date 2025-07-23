package keeper_test

import (
	"fmt"
	"testing"

	errorsmod "cosmossdk.io/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/suite"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	hlcoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T())
	ctx := app.NewContext(false)

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

func (suite *KeeperTestSuite) TestCreateDenomHLToken() {
	k := suite.App.DenomMetadataKeeper
	bankKeeper := suite.App.BankKeeper
	hlCoreK := suite.App.HyperCoreKeeper
	mailbox := util.GenerateHexAddress([20]byte{1}, 1, 1) // dummy value
	err := hlCoreK.Mailboxes.Set(suite.Ctx, mailbox.GetInternalId(), hlcoretypes.Mailbox{})
	suite.Require().NoError(err)

	warpS := warpkeeper.NewMsgServerImpl(suite.App.HyperWarpKeeper)

	signer := "dym1000000000000000000000000000000000000000000000000000000000000000"

	msg0 := warptypes.MsgCreateSyntheticToken{
		OriginMailbox: mailbox,
		Owner:         signer,
	}
	res0, err := warpS.CreateSyntheticToken(suite.Ctx, &msg0)
	suite.Require().NoError(err)

	metadata := suite.getDymMetadata()
	metadata.Base = fmt.Sprintf("hyperlane/%s", res0.Id.String())
	metadata.DenomUnits[0].Denom = metadata.Base

	msg1 := types.MsgRegisterHLTokenDenomMetadata{
		HlTokenId:     res0.Id,
		HlTokenOwner:  signer,
		TokenMetadata: metadata,
	}

	suite.Require().NoError(metadata.Validate())

	s := keeper.NewMsgServerImpl(k)
	_, err = s.RegisterHLTokenDenomMetadata(suite.Ctx, &msg1)
	suite.Require().NoError(err)

	denom, found := bankKeeper.GetDenomMetaData(suite.Ctx, metadata.Base)
	suite.Require().EqualValues(found, true)
	suite.Require().EqualValues(denom.Symbol, metadata.Symbol)
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
