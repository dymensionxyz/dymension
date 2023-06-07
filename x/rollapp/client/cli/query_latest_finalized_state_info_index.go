package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cobra"
)

func CmdShowLatestFinalizedStateIndex() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-finalized-state-index [rollapp-id]",
		Short: "Query the index of the last UpdateState that was finalized associated with the specified rollapp-id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			params := &types.QueryGetLatestFinalizedStateIndexRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.LatestFinalizedStateIndex(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
