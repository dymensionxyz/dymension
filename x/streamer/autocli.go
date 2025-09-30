package streamer

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              "dymensionxyz.dymension.streamer.Query",
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "PumpPressure",
					Use:       "pump-pressure",
					Short:     "Query pump pressure for all rollapps from all streams",
					Long:      "Returns how much DYM will be used for buying RA tokens if pump occurs for all rollapps from all streams.",
				},
				{
					RpcMethod: "PumpPressureByRollapp",
					Use:       "pump-pressure-by-rollapp [rollapp-id]",
					Short:     "Query pump pressure for a specific rollapp from all streams",
					Long:      "Returns how much DYM will be used for buying RA tokens if pump occurs for a specific rollapp from all streams.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "rollapp_id"},
					},
				},
				{
					RpcMethod: "PumpPressureByUser",
					Use:       "pump-pressure-by-user [address]",
					Short:     "Query pump pressure by user for all rollapps from all streams",
					Long:      "Returns how much pump pressure the user puts on RAs with their cast voting power for all rollapps from all streams.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
					},
				},
				{
					RpcMethod: "PumpPressureByUserByRollapp",
					Use:       "pump-pressure-by-user-by-rollapp [address] [rollapp-id]",
					Short:     "Query pump pressure by user for a specific rollapp from all streams",
					Long:      "Returns how much pump pressure the user puts on RAs with their cast voting power for a specific rollapp from all streams.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "address"},
						{ProtoField: "rollapp_id"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "dymensionxyz.dymension.streamer.Msg",
		},
	}
}