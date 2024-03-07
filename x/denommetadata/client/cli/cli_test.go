package cli_test

import (
	"testing"

	cli "github.com/dymensionxyz/dymension/v3/x/denommetadata/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

func TestGetAllDenomMetadata(t *testing.T) {
	desc, _ := cli.GetCmdAllDenomMetadata()
	tcs := map[string]osmocli.QueryCliTestCase[*types.AllDenomMetadataRequest]{
		"basic test": {
			Cmd:           "",
			ExpectedQuery: &types.AllDenomMetadataRequest{},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdDenomMetadataByID(t *testing.T) {
	desc, _ := cli.GetCmdDenomMetadataByID()
	tcs := map[string]osmocli.QueryCliTestCase[*types.DenomMetadataByIDRequest]{
		"basic test": {
			Cmd: "1", ExpectedQuery: &types.DenomMetadataByIDRequest{Id: 1},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdDenomMetadataByBase(t *testing.T) {
	desc, _ := cli.GetCmdDenomMetadataByBaseDenom()
	tcs := map[string]osmocli.QueryCliTestCase[*types.DenomMetadataByBaseDenomRequest]{
		"basic test": {
			Cmd: "uatom", ExpectedQuery: &types.DenomMetadataByBaseDenomRequest{BaseDenom: "uatom"},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdDenomMetadataBySymbol(t *testing.T) {
	desc, _ := cli.GetCmdDenomMetadataBySymbolDenom()
	tcs := map[string]osmocli.QueryCliTestCase[*types.DenomMetadataBySymbolDenomRequest]{
		"basic test": {
			Cmd: "ATOM", ExpectedQuery: &types.DenomMetadataBySymbolDenomRequest{SymbolDenom: "ATOM"},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdDenomMetadataByDisplay(t *testing.T) {
	desc, _ := cli.GetCmdDenomMetadataBySymbolDenom()
	tcs := map[string]osmocli.QueryCliTestCase[*types.DenomMetadataByDisplayDenomRequest]{
		"basic test": {
			Cmd: "atom", ExpectedQuery: &types.DenomMetadataByDisplayDenomRequest{DisplayDenom: "atom"},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
