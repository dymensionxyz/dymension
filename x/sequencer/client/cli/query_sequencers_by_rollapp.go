package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func CmdShowSequencersByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-sequencers-by-rollapp [rollapp-id]",
		Short: "shows a sequencers_by_rollapp",
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
