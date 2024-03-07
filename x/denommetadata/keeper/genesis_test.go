package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

// TestDenomMetadataExportGenesis tests export genesis command for the denommetadata module.
func TestDenomMetadataExportGenesis(t *testing.T) {
	// export genesis using default configurations
	// ensure resulting genesis params match default params
	app := apptesting.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	genesis := app.DenomMetadataKeeper.ExportGenesis(ctx)
	require.Equal(t, genesis.Params, types.DefaultGenesis().Params)
	require.Len(t, genesis.Denommetadatas, 0)

	token := types.TokenMetadata{
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
	metadataId, err := app.DenomMetadataKeeper.CreateDenomMetadata(ctx, token)
	require.NoError(t, err)

	// export genesis using default configurations
	// ensure resulting genesis params match default params
	genesis = app.DenomMetadataKeeper.ExportGenesis(ctx)
	require.Len(t, genesis.Denommetadatas, 1)

	// ensure the first denommetadata listed in the exported genesis explicitly matches expectation
	require.Equal(t, genesis.Denommetadatas[0], types.DenomMetadata{
		Id:            metadataId,
		TokenMetadata: token,
	})
}

// TestDenomMetadataInitGenesis takes a genesis state and tests initializing that genesis for the denommetadata module.
func TestDenomMetadataInitGenesis(t *testing.T) {
	app := apptesting.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	// checks that the default genesis parameters pass validation
	validateGenesis := types.DefaultGenesis().Params.Validate()
	require.NoError(t, validateGenesis)

	tokenmetadata := types.TokenMetadata{
		Name:        "Cosmos Hub Atom",
		Symbol:      "ATOM",
		Description: "The native staking token of the Cosmos Hub.",
		DenomUnits: []*types.DenomUnit{
			{Denom: "uatom", Exponent: uint32(0), Aliases: []string{"microatom"}},
			{Denom: "atom", Exponent: uint32(6), Aliases: nil},
			{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
		},
		Base:    "uatom",
		Display: "atom",
	}
	denommetadata := types.DenomMetadata{
		Id:            1,
		TokenMetadata: tokenmetadata,
	}

	// initialize genesis with specified parameter, the denom metadata created earlier, and lockable durations
	app.DenomMetadataKeeper.InitGenesis(ctx, types.GenesisState{
		Params:              types.Params{},
		Denommetadatas:      []types.DenomMetadata{denommetadata},
		LastDenommetadataId: 1,
	})

	// check that the denommetadata created earlier was initialized through initGenesis and still exists on chain
	denommetadatas := app.DenomMetadataKeeper.GetAllDenomMetadata(ctx)
	lastDenommetadataId := app.DenomMetadataKeeper.GetLastDenomMetadataID(ctx)
	require.Len(t, denommetadatas, 1)
	require.Equal(t, denommetadatas[0], denommetadata)
	require.Equal(t, lastDenommetadataId, uint64(1))
}
