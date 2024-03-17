package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"

	cli "github.com/dymensionxyz/dymension/v3/x/streamer/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

func TestGetCmdStreams(t *testing.T) {
	desc, _ := cli.GetCmdStreams()
	tcs := map[string]osmocli.QueryCliTestCase[*types.StreamsRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.StreamsRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdToDistributeCoins(t *testing.T) {
	desc, _ := cli.GetCmdToDistributeCoins()
	tcs := map[string]osmocli.QueryCliTestCase[*types.ModuleToDistributeCoinsRequest]{
		"basic test": {
			Cmd: "", ExpectedQuery: &types.ModuleToDistributeCoinsRequest{},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdStreamByID(t *testing.T) {
	desc, _ := cli.GetCmdStreamByID()
	tcs := map[string]osmocli.QueryCliTestCase[*types.StreamByIDRequest]{
		"basic test": {
			Cmd: "1", ExpectedQuery: &types.StreamByIDRequest{Id: 1},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdActiveStreams(t *testing.T) {
	desc, _ := cli.GetCmdActiveStreams()
	tcs := map[string]osmocli.QueryCliTestCase[*types.ActiveStreamsRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.ActiveStreamsRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}

func TestGetCmdUpcomingStreams(t *testing.T) {
	desc, _ := cli.GetCmdUpcomingStreams()
	tcs := map[string]osmocli.QueryCliTestCase[*types.UpcomingStreamsRequest]{
		"basic test": {
			Cmd: "--offset=2",
			ExpectedQuery: &types.UpcomingStreamsRequest{
				Pagination: &query.PageRequest{Key: []uint8{}, Offset: 2, Limit: 100},
			},
		},
	}
	osmocli.RunQueryTestCases(t, desc, tcs)
}
