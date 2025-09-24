package bridgingfee

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service:              "dymensionxyz.dymension.bridgingfee.Query",
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "QuoteFeePayment",
					Use:       "quote-fee-payment [hook-id] [token-id] [transfer-amount]",
					Short:     "Quote the fee payment required for a transfer",
					Long:      "Quote the fee payment required for a transfer through a specific hook for a given token and amount.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "hook_id"},
						{ProtoField: "token_id"},
						{ProtoField: "transfer_amount"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: "dymensionxyz.dymension.bridgingfee.Msg",
		},
	}
}
