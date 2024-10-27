package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdQueryRegisteredDenoms() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registered-denoms",
		Short: "Lists the registered denoms for a Rollapp",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			rollappID := args[0]

			res, err := queryClient.RegisteredDenoms(cmd.Context(), &types.QueryRegisteredDenomsRequest{
				RollappId: rollappID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
