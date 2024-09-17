package cli

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	FlagOmitApps = "omit-apps"
)

func CmdListRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Query all rollapps currently registered in the hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			omitApps, err := cmd.Flags().GetBool(FlagOmitApps)
			if err != nil {
				return err
			}

			params := &types.QueryAllRollappRequest{
				Pagination: pageReq,
				OmitApps:   omitApps,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.RollappAll(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool(FlagOmitApps, false, "Omit the list of apps associated with each rollapp")

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [rollapp-id]",
		Short:   "Query the rollapp associated with the specified rollapp-id",
		Args:    cobra.ExactArgs(1),
		Example: "dymd query rollapp show ROLLAPP_CHAIN_ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			omitApps, err := cmd.Flags().GetBool(FlagOmitApps)
			if err != nil {
				return err
			}

			params := &types.QueryGetRollappRequest{
				RollappId: args[0],
				OmitApps:  omitApps,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Rollapp(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool(FlagOmitApps, false, "Omit the list of apps associated with the rollapp")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
