package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/sequencer/types"
	"github.com/spf13/cobra"
)

func CmdListScheduler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-scheduler",
		Short: "list all scheduler",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllSchedulerRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.SchedulerAll(context.Background(), params)
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

func CmdShowScheduler() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-scheduler [sequencer-address]",
		Short: "shows a scheduler",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argSequencerAddress := args[0]

			params := &types.QueryGetSchedulerRequest{
				SequencerAddress: argSequencerAddress,
			}

			res, err := queryClient.Scheduler(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
