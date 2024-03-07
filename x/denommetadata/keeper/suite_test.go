package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	"testing"

	"github.com/stretchr/testify/suite"

	keeper "github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	"github.com/osmosis-labs/osmosis/v15/app/apptesting"
)

var (
	defaultTokenMetadata = types.TokenMetadata{
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*types.DenomUnit{
			{Denom: "uatom", Exponent: uint32(0), Aliases: []string{"microatom"}},
			{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
			{Denom: "atom", Exponent: uint32(6), Aliases: nil},
		},
		Base:    "uatom",
		Display: "atom",
	}
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
	querier keeper.Querier
}

// SetupTest sets streamer parameters from the suite's context
func (suite *KeeperTestSuite) SetupTest() {
	suite.Setup()
	suite.querier = keeper.NewQuerier(suite.App.DenomMetadataKeeper)

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// CreateStream creates a stream struct given the required params.
func (suite *KeeperTestSuite) CreateMetadata(record types.TokenMetadata) (uint64, *types.DenomMetadata) {
	denomID, err := suite.App.DenomMetadataKeeper.CreateDenomMetadata(suite.Ctx, record)
	suite.Require().NoError(err)
	denomMetadata, err := suite.App.DenomMetadataKeeper.GetDenomMetadataByID(suite.Ctx, denomID)
	suite.Require().NoError(err)
	return denomID, denomMetadata
}

func (suite *KeeperTestSuite) CreateDefaultDenomMetadata() (uint64, *types.DenomMetadata) {
	return suite.CreateMetadata(defaultTokenMetadata)
}

func (suite *KeeperTestSuite) ExpectedDefaultDenomMetadata(denomID uint64) types.DenomMetadata {

	return types.DenomMetadata{
		Id:            denomID,
		TokenMetadata: defaultTokenMetadata,
	}

}
