package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/spf13/cobra"
)

func CmdShowLatestHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-height [rollapp-id]",
		Short: "Query the last height of the last UpdateState associated with the specified rollapp-id.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			argFinalized, err := cmd.Flags().GetBool(FlagFinalized)
			if err != nil {
				return err
			}

			req := &types.QueryGetLatestHeightRequest{
				RollappId: argRollappId,
				Finalized: argFinalized,
			}

			res, err := queryClient.LatestHeight(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool(FlagFinalized, false, "Indicates whether to return the latest finalized state index")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
