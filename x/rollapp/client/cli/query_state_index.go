package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cobra"
)

func CmdListStateIndex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-state-index",
		Short: "list all state_index",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllStateIndexRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.StateIndexAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowStateIndex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-state-index [rollapp-id]",
		Short: "shows a state_index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			params := &types.QueryGetStateIndexRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.StateIndex(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
