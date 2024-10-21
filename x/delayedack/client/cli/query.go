package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
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
	cmd.AddCommand(CmdGetPacketsByType())
	cmd.AddCommand(CmdGetPendingPacketsByReceiver())

	return cmd
}

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "shows the parameters of the module",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetPacketsByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packets-by-rollapp rollapp-id [status] [type]",
		Short: "Get packets by rollapp-id",
		Long: `Get packets by rollapp-id. Can filter by status (pending/finalized/reverted) and by type (recv/ack/timeout)
		Example:
		packets rollapp1
		packets rollapp1 PENDING
		packets rollapp1 PENDING RECV`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rollappId := args[0]

			req := &types.QueryRollappPacketsRequest{
				RollappId: rollappId,
				Status:    commontypes.Status_PENDING,          // get pending packets by default
				Type:      commontypes.RollappPacket_UNDEFINED, // must specify, as '0' is a valid type
			}

			if len(args) > 1 {
				statusStr := strings.ToUpper(args[1])
				status, ok := commontypes.Status_value[statusStr]
				if !ok {
					// Handle error: statusStr is not a valid commontypes.Status
					return fmt.Errorf("invalid status: %s", statusStr)
				}
				req.Status = commontypes.Status(status)
			}

			if len(args) > 2 {
				typeStr := strings.ToUpper(args[2])
				if !strings.HasPrefix(typeStr, "ON_") {
					typeStr = "ON_" + typeStr
				}
				dtype, ok := commontypes.RollappPacket_Type_value[typeStr]
				if !ok {
					// Handle error: typeStr is not a valid commontypes.Type
					return fmt.Errorf("invalid type: %s", typeStr)
				}
				req.Type = commontypes.RollappPacket_Type(dtype)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetPackets(cmd.Context(), req)
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
		Use:   "packets-by-status status [type]",
		Short: "Get packets by status",
		Long: `Get packets by status (pending/finalized/reverted). Can filter by type (recv/ack/timeout)
		Example:
		packets-by-status pending
		packets-by-status finalized recv`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			statusStr := strings.ToUpper(args[0])
			status, ok := commontypes.Status_value[statusStr]
			if !ok {
				return fmt.Errorf("invalid status: %s", statusStr)
			}

			req := &types.QueryRollappPacketsRequest{
				Status: commontypes.Status(status),
				Type:   commontypes.RollappPacket_UNDEFINED, // must specify, as '0' is a valid type
			}

			if len(args) > 1 {
				typeStr := strings.ToUpper(args[1])
				if !strings.HasPrefix(typeStr, "ON_") {
					typeStr = "ON_" + typeStr
				}
				dtype, ok := commontypes.RollappPacket_Type_value[typeStr]
				if !ok {
					return fmt.Errorf("invalid type: %s", typeStr)
				}
				req.Type = commontypes.RollappPacket_Type(dtype)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetPackets(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetPacketsByType() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "packets-by-type type",
		Short: "Get pending packets by type",
		Long: `Get pending packets by type. Can filter by type (recv/ack/timeout)
		Example:
		packets-by-type on_recv
		packets-by-type on_timeout`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			typeStr := strings.ToUpper(args[0])

			if !strings.HasPrefix(typeStr, "ON_") {
				typeStr = "ON_" + typeStr
			}

			dtype, ok := commontypes.RollappPacket_Type_value[typeStr]
			if !ok {
				return fmt.Errorf("invalid type: %s", typeStr)
			}

			req := &types.QueryRollappPacketsRequest{
				Type:   commontypes.RollappPacket_Type(dtype),
				Status: commontypes.Status_PENDING, // get pending packets by default
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetPackets(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetPendingPacketsByReceiver() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending-packets-by-receiver [rollapp-id] [receiver]",
		Short: "Get pending packets by receiver",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetPendingPacketsByReceiver(cmd.Context(), &types.QueryPendingPacketsByReceiverRequest{
				RollappId:  args[0],
				Receiver:   args[1],
				Pagination: nil, // TODO: handle pagination
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
