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
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingStreams)
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
			Long: `This command allows you to query the coins that are scheduled to be distributed.
		It returns a list of coins with their denominations and amounts.`},
		&types.ModuleToDistributeCoinsRequest{}
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
			Long: `This command allows you to query all active streams.
		An active stream is a stream that is currently in progress.
		The command returns a list of active streams with their details.`},
		&types.ActiveStreamsRequest{}
}

// GetCmdUpcomingStreams returns scheduled streams.
func GetCmdUpcomingStreams() (*osmocli.QueryDescriptor, *types.UpcomingStreamsRequest) {
	return &osmocli.QueryDescriptor{
			Use:   "upcoming-streams",
			Short: "Query upcoming streams",
			Long: `This command allows you to query all upcoming streams.
		An upcoming stream is a stream that is scheduled to start in the future.
		The command returns a list of upcoming streams with their details, including the start time, end time, and the coins it contains.`},
		&types.UpcomingStreamsRequest{}
}
