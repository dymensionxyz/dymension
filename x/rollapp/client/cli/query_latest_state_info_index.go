package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cobra"
)

func CmdListLatestStateInfoIndex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-state-index",
		Short: "list all latest_state_info_index",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllLatestStateInfoIndexRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.LatestStateInfoIndexAll(context.Background(), params)
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

func CmdShowLatestStateInfoIndex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-latest-state-index [rollapp-id]",
		Short: "shows the latest_state_info_index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			params := &types.QueryGetLatestStateInfoIndexRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.LatestStateInfoIndex(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
