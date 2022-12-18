package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdLatestFinalizedStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-finalized-state-info [rollapp-id]",
		Short: "query the latest StateInfo that was finalized",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqRollappId := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryLatestFinalizedStateInfoRequest{

				RollappId: reqRollappId,
			}

			res, err := queryClient.LatestFinalizedStateInfo(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
