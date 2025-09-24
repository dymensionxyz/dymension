package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryFeeHooks(),
		CmdQueryFeeHook(),
		CmdQueryAggregationHooks(),
		CmdQueryAggregationHook(),
	)

	return cmd
}

// CmdQueryFeeHooks implements the query fee-hooks command.
func CmdQueryFeeHooks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-hooks",
		Short: "Query all fee hooks",
		Long:  "Query all fee hooks with optional pagination.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.FeeHooks(context.Background(), &types.QueryFeeHooksRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "fee-hooks")

	return cmd
}

// CmdQueryFeeHook implements the query fee-hook command.
func CmdQueryFeeHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee-hook [hook-id]",
		Short: "Query a fee hook by ID",
		Long:  "Query the details of a specific fee hook using its unique identifier.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.FeeHook(context.Background(), &types.QueryFeeHookRequest{
				Id: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryAggregationHooks implements the query aggregation-hooks command.
func CmdQueryAggregationHooks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregation-hooks",
		Short: "Query all aggregation hooks",
		Long:  "Query all aggregation hooks with optional pagination.",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AggregationHooks(context.Background(), &types.QueryAggregationHooksRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "aggregation-hooks")

	return cmd
}

// CmdQueryAggregationHook implements the query aggregation-hook command.
func CmdQueryAggregationHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aggregation-hook [hook-id]",
		Short: "Query an aggregation hook by ID",
		Long:  "Query the details of a specific aggregation hook using its unique identifier.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.AggregationHook(context.Background(), &types.QueryAggregationHookRequest{
				Id: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
