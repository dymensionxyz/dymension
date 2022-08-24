package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListRollappStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-rollapp-state-info",
		Short: "list all rollapp_state_info",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllRollappStateInfoRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.RollappStateInfoAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowRollappStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-rollapp-state-info [rollapp-id] [state-index]",
		Short: "shows a rollapp_state_info",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]
			argStateIndex, err := cast.ToUint64E(args[1])
			if err != nil {
				return err
			}

			params := &types.QueryGetRollappStateInfoRequest{
				RollappId:  argRollappId,
				StateIndex: argStateIndex,
			}

			res, err := queryClient.RollappStateInfo(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
