package cli

import (
	"context"
	"fmt"
	"strings"

	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group eibc queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdGetPacketsByRollapp())
	cmd.AddCommand(CmdGetPacketsByStatus())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Define the map
var statusMap = map[string]commontypes.Status{
	"PENDING":   commontypes.Status_PENDING,
	"FINALIZED": commontypes.Status_FINALIZED,
	"REVERTED":  commontypes.Status_REVERTED,
}

func CmdGetPacketsByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packets-by-rollapp rollapp-id [status]",
		Short: "get packets by rollapp-id",
		Long: `get packets by rollapp-id. Can filter by status (pending/finalized/reverted) 
		Example:
		packets rollapp1 PENDING
		packets rollapp1
		packets PENDING`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			rollappId := args[0]

			req := &types.QueryRollappPacketsRequest{
				RollappId: rollappId,
				Status:    commontypes.Status_PENDING, // get pending packets by default
			}

			if len(args) > 1 {
				// Use the map to convert a string to an enum
				statusStr := strings.ToUpper(args[1])
				status, ok := statusMap[statusStr]
				if !ok {
					// Handle error: statusStr is not a valid commontypes.Status
					return fmt.Errorf("invalid status: %s", statusStr)
				}
				req.Status = status
			}

			res, err := queryClient.GetPackets(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetPacketsByStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packets-by-status status",
		Short: "get packets by status",
		Long: `get packets by status. Can filter by status (pending/finalized/reverted)
		Example:
		packets-by-status pending
		packets-by-status finalized`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			statusStr := strings.ToUpper(args[0])
			status, ok := statusMap[statusStr]
			if !ok {
				return fmt.Errorf("invalid status: %s", statusStr)
			}

			req := &types.QueryRollappPacketsRequest{
				Status: status,
			}

			res, err := queryClient.GetPackets(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
