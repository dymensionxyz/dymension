package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdQueryObsoleteDRSVersions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "obsolete-drs-versions",
		Short: "shows the list of obsolete DRS versions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.ObsoleteDRSVersions(cmd.Context(), &types.QueryObsoleteDRSVersionsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
