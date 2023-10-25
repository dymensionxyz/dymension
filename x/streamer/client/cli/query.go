package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group streamer queries under a subcommand
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	qcGetter := types.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdStreams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdToDistributeCoins)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdStreamByID)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveStreams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveStreamsPerDenom)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingStreams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingStreamsPerDenom)

	return cmd
}

// GetCmdStreams returns all available streams.
func GetCmdStreams() (*osmocli.QueryDescriptor, *types.StreamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "streams",
		Short: "Query all available streams",
		Long:  "{{.Short}}",
	}, &types.StreamsRequest{}
}

// GetCmdToDistributeCoins returns coins that are going to be distributed.
func GetCmdToDistributeCoins() (*osmocli.QueryDescriptor, *types.ModuleToDistributeCoinsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "to-distribute-coins",
		Short: "Query coins that is going to be distributed",
		Long:  `{{.Short}}`}, &types.ModuleToDistributeCoinsRequest{}
}

// GetCmdStreamByID returns a stream by ID.
func GetCmdStreamByID() (*osmocli.QueryDescriptor, *types.StreamByIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "stream-by-id [id]",
		Short: "Query stream by id.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} stream-by-id 1
`}, &types.StreamByIDRequest{}
}

// GetCmdActiveStreams returns active streams.
func GetCmdActiveStreams() (*osmocli.QueryDescriptor, *types.ActiveStreamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "active-streams",
		Short: "Query active streams",
		Long:  `{{.Short}}`}, &types.ActiveStreamsRequest{}
}

// GetCmdActiveStreamsPerDenom returns active streams for a specified denom.
func GetCmdActiveStreamsPerDenom() (*osmocli.QueryDescriptor, *types.ActiveStreamsPerDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "active-streams-per-den [den]denom [denom]",
		Short: "Query active streams per denom",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} active-streams-per-denom gamm/pool/1`}, &types.ActiveStreamsPerDenomRequest{}
}

// GetCmdUpcomingStreams returns scheduled streams.
func GetCmdUpcomingStreams() (*osmocli.QueryDescriptor, *types.UpcomingStreamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "upcoming-streams",
		Short: "Query upcoming streams",
		Long:  `{{.Short}}`}, &types.UpcomingStreamsRequest{}
}

// GetCmdUpcomingStreamsPerDenom returns scheduled streams for specified denom..
func GetCmdUpcomingStreamsPerDenom() (*osmocli.QueryDescriptor, *types.UpcomingStreamsPerDenomRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "upcoming-streams-per-denom [denom]",
		Short: "Query scheduled streams per denom",
		Long:  `{{.Short}}`}, &types.UpcomingStreamsPerDenomRequest{}
}

func contains(s []uint64, value uint64) bool {
	for _, v := range s {
		if v == value {
			return true
		}
	}

	return false
}
