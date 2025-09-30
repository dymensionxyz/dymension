package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetQueryCmd returns the query commands for this module.
func GetQueryCmd() *cobra.Command {
	// group streamer queries under a subcommand
	cmd := osmocli.QueryIndexCmd(types.ModuleName)
	qcGetter := types.NewQueryClient
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdParams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdStreams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdToDistributeCoins)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdStreamByID)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdActiveStreams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdUpcomingStreams)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdPumpPressure)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdPumpPressureByRollapp)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdPumpPressureByUser)
	osmocli.AddQueryCmd(cmd, qcGetter, GetCmdPumpPressureByUserByRollapp)
	return cmd
}

// GetCmdParams returns the streamer module parameters.
func GetCmdParams() (*osmocli.QueryDescriptor, *types.ParamsRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "params",
		Short: "Query streamer module parameters",
		Long:  "Query the current parameters of the streamer module",
	}, &types.ParamsRequest{}
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
		It returns a list of coins with their denominations and amounts.`,
		},
		&types.ModuleToDistributeCoinsRequest{}
}

// GetCmdStreamByID returns a stream by ID.
func GetCmdStreamByID() (*osmocli.QueryDescriptor, *types.StreamByIDRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "stream-by-id [id]",
		Short: "Query stream by id.",
		Long: `{{.Short}}{{.ExampleHeader}}
{{.CommandPrefix}} stream-by-id 1
`,
	}, &types.StreamByIDRequest{}
}

// GetCmdActiveStreams returns active streams.
func GetCmdActiveStreams() (*osmocli.QueryDescriptor, *types.ActiveStreamsRequest) {
	return &osmocli.QueryDescriptor{
			Use:   "active-streams",
			Short: "Query active streams",
			Long: `This command allows you to query all active streams.
		An active stream is a stream that is currently in progress.
		The command returns a list of active streams with their details.`,
		},
		&types.ActiveStreamsRequest{}
}

// GetCmdUpcomingStreams returns scheduled streams.
func GetCmdUpcomingStreams() (*osmocli.QueryDescriptor, *types.UpcomingStreamsRequest) {
	return &osmocli.QueryDescriptor{
			Use:   "upcoming-streams",
			Short: "Query upcoming streams",
			Long: `This command allows you to query all upcoming streams.
		An upcoming stream is a stream that is scheduled to start in the future.
		The command returns a list of upcoming streams with their details, including the start time, end time, and the coins it contains.`,
		},
		&types.UpcomingStreamsRequest{}
}

// GetCmdPumpPressure returns pump pressure for all rollapps.
func GetCmdPumpPressure() (*osmocli.QueryDescriptor, *types.PumpPressureRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pump-pressure",
		Short: "Query pump pressure for all rollapps from all streams",
		Long:  "Returns how much DYM will be used for buying RA tokens if pump occurs for all rollapps from all streams.",
	}, &types.PumpPressureRequest{}
}

// GetCmdPumpPressureByRollapp returns pump pressure for a specific rollapp.
func GetCmdPumpPressureByRollapp() (*osmocli.QueryDescriptor, *types.PumpPressureByRollappRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pump-pressure-by-rollapp [rollapp-id]",
		Short: "Query pump pressure for a specific rollapp from all streams",
		Long:  "Returns how much DYM will be used for buying RA tokens if pump occurs for a specific rollapp from all streams.",
	}, &types.PumpPressureByRollappRequest{}
}

// GetCmdPumpPressureByUser returns pump pressure by user for all rollapps.
func GetCmdPumpPressureByUser() (*osmocli.QueryDescriptor, *types.PumpPressureByUserRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pump-pressure-by-user [address]",
		Short: "Query pump pressure by user for all rollapps from all streams",
		Long:  "Returns how much pump pressure the user puts on RAs with their cast voting power for all rollapps from all streams.",
	}, &types.PumpPressureByUserRequest{}
}

// GetCmdPumpPressureByUserByRollapp returns pump pressure by user for a specific rollapp.
func GetCmdPumpPressureByUserByRollapp() (*osmocli.QueryDescriptor, *types.PumpPressureByUserByRollappRequest) {
	return &osmocli.QueryDescriptor{
		Use:   "pump-pressure-by-user-by-rollapp [address] [rollapp-id]",
		Short: "Query pump pressure by user for a specific rollapp from all streams",
		Long:  "Returns how much pump pressure the user puts on RAs with their cast voting power for a specific rollapp from all streams.",
	}, &types.PumpPressureByUserByRollappRequest{}
}
