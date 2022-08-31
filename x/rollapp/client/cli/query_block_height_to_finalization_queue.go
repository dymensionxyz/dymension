package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListBlockHeightToFinalizationQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-block-height-to-finalization-queue",
		Short: "list all block_height_to_finalization_queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllBlockHeightToFinalizationQueueRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.BlockHeightToFinalizationQueueAll(context.Background(), params)
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

func CmdShowBlockHeightToFinalizationQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-block-height-to-finalization-queue [finalization-height]",
		Short: "shows a block_height_to_finalization_queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argFinalizationHeight, err := cast.ToUint64E(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryGetBlockHeightToFinalizationQueueRequest{
				FinalizationHeight: argFinalizationHeight,
			}

			res, err := queryClient.BlockHeightToFinalizationQueue(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
