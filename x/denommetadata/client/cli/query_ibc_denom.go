package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/spf13/cobra"
)

func CmdIBCDenomBaseOnDenomTrace() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ibc-denom [port-id-1]/[channel-id-1]/.../[port-id-n]/[channel-id-n]/[denom]",
		Short: "Get IBC denom base on a denom trace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			denomTrace := args[0]

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.IBCDenomByDenomTrace(context.Background(), &types.QueryGetIBCDenomByDenomTraceRequest{
				DenomTrace: denomTrace,
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
