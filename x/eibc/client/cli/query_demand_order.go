package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func CmdGetDemandOrderById() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "demand-order [order-id]",
		Short: "Query a demand order by its ID",
		Long: `Query a specific eIBC demand order by its unique identifier.
The command searches for the order across all statuses (pending, finalized).

Example:
  dymd query eibc demand-order 0d784ae938d0e00c2a047429da5968b4c3437ac8cfa8130204914df4d3430628`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DemandOrderById(cmd.Context(), &types.QueryGetDemandOrderRequest{
				Id: args[0],
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
