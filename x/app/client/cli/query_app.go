package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/app/types"
)

func CmdListApp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Query all apps currently registered in the hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllAppRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.AppAll(cmd.Context(), params)
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

func CmdShowApp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name] [rollapp-id]",
		Short:   "Query the app associated with the specified name",
		Args:    cobra.ExactArgs(2),
		Example: "dymd query app show app1 rollapp_1234-1",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetAppRequest{
				Name:      args[0],
				RollappId: args[1],
			}

			res, err := queryClient.App(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
