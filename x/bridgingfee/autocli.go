package bridgingfee

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "dymensionxyz.dymension.bridgingfee.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "FeeHook",
					Use:       "fee-hook [hook-id]",
					Short:     "Query a fee hook by ID",
					Long:      "Query the details of a specific fee hook using its unique identifier.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "FeeHooks",
					Use:       "fee-hooks",
					Short:     "Query all fee hooks",
					Long:      "Query all fee hooks with optional pagination.",
				},
				{
					RpcMethod: "AggregationHook",
					Use:       "aggregation-hook [hook-id]",
					Short:     "Query an aggregation hook by ID",
					Long:      "Query the details of a specific aggregation hook using its unique identifier.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
				},
				{
					RpcMethod: "AggregationHooks",
					Use:       "aggregation-hooks",
					Short:     "Query all aggregation hooks",
					Long:      "Query all aggregation hooks with optional pagination.",
				},
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
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateBridgingFeeHook",
					Use:       "create-fee-hook [fees...]",
					Short:     "Create a new bridging fee hook",
					Long:      "Create a new fee hook that charges fees for token transfers across bridges. Fees should be provided as JSON objects.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "fees", Varargs: true},
					},
				},
				{
					RpcMethod: "SetBridgingFeeHook",
					Use:       "set-fee-hook [hook-id]",
					Short:     "Update an existing bridging fee hook",
					Long:      "Update the configuration of an existing fee hook, including fees, ownership, or other settings.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"fees": {
							Name:  "update-fees",
							Usage: "Fee configuration for each token (JSON format: {\"tokenId\":\"0x...\",\"inboundFee\":\"0.01\",\"outboundFee\":\"0.02\"}). Can be repeated for multiple tokens.",
						},
						"new_owner": {
							Name:  "new-owner",
							Usage: "Transfer ownership to this address",
						},
						"renounce_ownership": {
							Name:  "renounce-ownership",
							Usage: "Renounce ownership of the hook",
						},
					},
				},
				{
					RpcMethod: "CreateAggregationHook",
					Use:       "create-aggregation-hook",
					Short:     "Create a new aggregation hook",
					Long:      "Create a new aggregation hook that combines multiple sub-hooks to execute them sequentially.",
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"hook_ids": {
							Name: "hook-ids",
						},
					},
				},
				{
					RpcMethod: "SetAggregationHook",
					Use:       "set-aggregation-hook [hook-id]",
					Short:     "Update an existing aggregation hook",
					Long:      "Update the configuration of an existing aggregation hook, including the list of sub-hooks or ownership settings.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "id"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"hook_ids": {
							Name: "hook-ids",
						},
						"new_owner": {
							Name: "new-owner",
						},
						"renounce_ownership": {
							Name: "renounce-ownership",
						},
					},
				},
			},
		},
	}
}
