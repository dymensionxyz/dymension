package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/spf13/cobra"
)

func CmdListDemandOrdersByStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-demand-orders [status]",
		Short: "List all demand orders with a specific status",
		Long:  `Query demand orders filtered by status. Examples of status include "pending", "finalized", and "reverted".`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			status := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.DemandOrdersByStatus(context.Background(), &types.QueryDemandOrdersByStatusRequest{
				Status: status,
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
