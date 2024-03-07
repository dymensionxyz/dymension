package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

var _ = suite.TestingSuite(nil)

// TestGRPCDenomMetadataByID tests querying denommetadata via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCDenomMetadataByID() {
	suite.SetupTest()

	// create denommetadata
	denomID, _ := suite.CreateDefaultDenomMetadata()

	// ensure that querying for a stream with an ID that doesn't exist returns an error.
	res, err := suite.querier.DenomMetadataByID(sdk.WrapSDKContext(suite.Ctx), &types.DenomMetadataByIDRequest{Id: 1000})
	suite.Require().Error(err)
	suite.Require().Equal(res, (*types.DenomMetadataByIDResponse)(nil))

	// check that querying a stream with an ID that exists returns the stream.
	res, err = suite.querier.DenomMetadataByID(sdk.WrapSDKContext(suite.Ctx), &types.DenomMetadataByIDRequest{Id: denomID})
	suite.Require().NoError(err)
	suite.Require().NotEqual(res.Metadata, nil)

	expectedMetadata := suite.ExpectedDefaultDenomMetadata(denomID)
	suite.Require().Equal(res.Metadata.String(), expectedMetadata.String())
}

// TestGRPCStreams tests querying upcoming and active streams via gRPC returns the correct response.
func (suite *KeeperTestSuite) TestGRPCDenomMetadatas() {
	suite.SetupTest()

	// ensure initially querying  returns no metadata
	res, err := suite.querier.AllDenomMetadata(sdk.WrapSDKContext(suite.Ctx), &types.AllDenomMetadataRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 0)

	// create denoom
	denomID, _ := suite.CreateDefaultDenomMetadata()

	// query streams again, but this time expect the stream created earlier in the response
	res, err = suite.querier.AllDenomMetadata(sdk.WrapSDKContext(suite.Ctx), &types.AllDenomMetadataRequest{})
	suite.Require().NoError(err)
	suite.Require().Len(res.Data, 1)
	expectedStream := suite.ExpectedDefaultDenomMetadata(denomID)
	suite.Require().Equal(res.Data[0].String(), expectedStream.String())

}

//TODO (srene): add tests by denom extra params
