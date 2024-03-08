package keeper_test

import (
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/stretchr/testify/suite"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestCreateMetadata() {
	tests := []struct {
		name      string
		metadata  types.TokenMetadata
		expectErr bool
	}{
		{
			name: "happy flow",
			metadata: types.TokenMetadata{
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
			},
			expectErr: false,
		},
		{
			name: "wrong denom unit",
			metadata: types.TokenMetadata{
				Name:        "Cosmos Hub Atom",
				Symbol:      "ATOM",
				Description: "The native staking token of the Cosmos Hub.",
				DenomUnits: []*types.DenomUnit{
					{Denom: "uatae", Exponent: uint32(0), Aliases: []string{"microatom"}},
					{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
					{Denom: "atom", Exponent: uint32(6), Aliases: nil},
				},
				Base:    "uatom",
				Display: "atom",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		suite.SetupTest()
		_, err := suite.App.DenomMetadataKeeper.CreateDenomMetadata(suite.Ctx, tc.metadata)
		if tc.expectErr {
			suite.Require().Error(err, tc.name)
		} else {
			suite.Require().NoError(err, tc.name)
		}
	}
}

func (suite *KeeperTestSuite) TestCreateExistingMetadata() {

	suite.SetupTest()

	// create denoom
	denomID, _ := suite.CreateDefaultDenomMetadata()
	expectedStream := suite.ExpectedDefaultDenomMetadata(denomID)

	_, _ = suite.CreateMetadata(expectedStream.TokenMetadata)

	err := suite.App.DenomMetadataKeeper.CheckExistingMetadata(suite.Ctx, expectedStream.TokenMetadata)
	suite.Require().Error(err)

}
