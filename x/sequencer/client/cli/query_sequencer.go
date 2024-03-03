package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/spf13/cobra"
)

func CmdListSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-sequencer",
		Short: "list all sequencer",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QuerySequencersRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.Sequencers(context.Background(), params)
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

func CmdShowSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-sequencer [sequencer-address]",
		Short: "shows a sequencer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argSequencerAddress := args[0]

			params := &types.QueryGetSequencerRequest{
				SequencerAddress: argSequencerAddress,
			}

			res, err := queryClient.Sequencer(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
