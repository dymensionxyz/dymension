package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGetStateInfoByHeight() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-state-info-by-height [rollapp-id] [height]",
		Short: "query StateInfo of the batch containing this rollapp block height",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqRollappId := args[0]
			reqHeight, err := cast.ToUint64E(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetStateInfoByHeightRequest{

				RollappId: reqRollappId,
				Height:    reqHeight,
			}

			res, err := queryClient.GetStateInfoByHeight(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
