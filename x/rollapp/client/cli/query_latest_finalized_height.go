package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdShowLatestFinalizedHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-finalized-height [rollapp-id]",
		Short: "Query the last finalized height associated with the specified rollapp-id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			req := &types.QueryGetLatestFinalizedHeightRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.LatestFinalizedHeight(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
