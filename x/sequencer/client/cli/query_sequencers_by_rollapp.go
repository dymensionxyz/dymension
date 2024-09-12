package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func CmdShowSequencersByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-sequencers-by-rollapp [rollapp-id]",
		Short: "shows the sequencers of a specific rollapp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argRollappId := args[0]

			params := &types.QueryGetSequencersByRollappRequest{
				RollappId: argRollappId,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.SequencersByRollapp(cmd.Context(), params)
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
			argRollappId := args[0]

			params := &types.QueryGetProposerByRollappRequest{
				RollappId: argRollappId,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetProposerByRollapp(cmd.Context(), params)
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
			argRollappId := args[0]

			params := &types.QueryGetNextProposerByRollappRequest{
				RollappId: argRollappId,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.GetNextProposerByRollapp(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdGetAllProposers() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-proposer",
		Short: "List all proposers",
		RunE: func(cmd *cobra.Command, args []string) error {
			params := &types.QueryProposersRequest{}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Proposers(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
