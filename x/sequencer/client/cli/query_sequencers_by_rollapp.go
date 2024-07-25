package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/spf13/cobra"
)

func CmdShowSequencersByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-sequencers-by-rollapp [rollapp-id]",
		Short: "shows the sequencers of a specific rollapp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			params := &types.QueryGetSequencersByRollappRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.SequencersByRollapp(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetProposerByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposer [rollapp-id]",
		Short: "Get the current proposer by rollapp ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			argRollappId := args[0]

			params := &types.QueryGetProposerByRollappRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.GetProposerByRollapp(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdGetNextProposerByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next-proposer [rollapp-id]",
		Short: "Get the next proposer by rollapp ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			argRollappId := args[0]

			params := &types.QueryGetNextProposerByRollappRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.GetNextProposerByRollapp(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
